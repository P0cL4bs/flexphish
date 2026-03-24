package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"flexphish/internal/config"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/template"
	"flexphish/pkg/utils"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

type TemplateHandler struct {
	repo    template.TemplateRepository
	repoCam campaign.Repository
}

type templateRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type deleteTemplateRequest struct {
	Filename string `json:"filename"`
}

type cloneTemplateRequest struct {
	NewFilename string  `json:"new_filename"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

const maxTemplateImportZipSize int64 = 50 * 1024 * 1024

var validHTTPMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodHead:    {},
	http.MethodOptions: {},
}

func buildTemplateDir(filename string) (string, error) {

	name := strings.TrimSuffix(filename, ".yaml")

	valid := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !valid.MatchString(name) {
		return "", errors.New("invalid filename")
	}

	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, "-", "_")

	return slug, nil
}

func validateTemplatePayload(tpl *template.Template) error {
	if strings.TrimSpace(tpl.Info.Name) == "" {
		return errors.New("info.name is required")
	}

	if strings.TrimSpace(tpl.Info.Author) == "" {
		return errors.New("info.author is required")
	}

	if len(tpl.Steps) == 0 {
		return errors.New("at least one step is required")
	}

	stepIDs := make(map[string]struct{}, len(tpl.Steps))

	for i, step := range tpl.Steps {
		pos := i + 1

		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("steps[%d].id is required", pos)
		}

		if _, exists := stepIDs[step.ID]; exists {
			return fmt.Errorf("steps[%d].id is duplicated", pos)
		}
		stepIDs[step.ID] = struct{}{}

		if strings.TrimSpace(step.Title) == "" {
			return fmt.Errorf("steps[%d].title is required", pos)
		}

		if strings.TrimSpace(step.Path) == "" {
			return fmt.Errorf("steps[%d].path is required", pos)
		}

		if !strings.HasPrefix(step.Path, "/") {
			return fmt.Errorf("steps[%d].path must start with /", pos)
		}

		if strings.TrimSpace(step.Method) == "" {
			return fmt.Errorf("steps[%d].method is required", pos)
		}

		method := strings.ToUpper(strings.TrimSpace(step.Method))
		if _, ok := validHTTPMethods[method]; !ok {
			return fmt.Errorf("steps[%d].method is invalid", pos)
		}

		if strings.TrimSpace(step.TemplateFile) == "" {
			return fmt.Errorf("steps[%d].template_file is required", pos)
		}

		if !strings.HasSuffix(strings.ToLower(step.TemplateFile), ".html") {
			return fmt.Errorf("steps[%d].template_file must be .html", pos)
		}

		if strings.Contains(step.TemplateFile, "..") {
			return fmt.Errorf("steps[%d].template_file is invalid", pos)
		}

		for j, field := range step.Capture.Fields {
			fieldPos := j + 1
			if strings.TrimSpace(field.Name) == "" {
				return fmt.Errorf("steps[%d].capture.fields[%d].name is required", pos, fieldPos)
			}
		}
	}

	for i, step := range tpl.Steps {
		if strings.TrimSpace(step.Next) == "" {
			continue
		}

		if _, ok := stepIDs[step.Next]; !ok {
			return fmt.Errorf("steps[%d].next references unknown step id", i+1)
		}
	}

	return nil
}

func NewTemplateHandler(repo template.TemplateRepository, camrepo campaign.Repository) *TemplateHandler {
	return &TemplateHandler{
		repo:    repo,
		repoCam: camrepo,
	}
}

func (h *TemplateHandler) ImportZip(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxTemplateImportZipSize); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid multipart form",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "file is required",
		})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "file must be a .zip",
		})
		return
	}

	limited := io.LimitReader(file, maxTemplateImportZipSize+1)
	zipBytes, err := io.ReadAll(limited)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "failed to read zip file",
		})
		return
	}
	if int64(len(zipBytes)) > maxTemplateImportZipSize {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "zip file too large",
		})
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid zip file",
		})
		return
	}

	yamlName := ""
	yamlContent := ""
	assets := make(map[string][]byte)

	for _, f := range zr.File {
		name := filepath.ToSlash(strings.TrimSpace(f.Name))
		if name == "" || strings.HasPrefix(name, "__MACOSX/") || strings.HasSuffix(name, "/") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" {
			if yamlName != "" {
				JSONResponse(w, http.StatusBadRequest, map[string]string{
					"error": "zip must contain exactly one .yaml file",
				})
				return
			}

			rc, openErr := f.Open()
			if openErr != nil {
				JSONResponse(w, http.StatusBadRequest, map[string]string{
					"error": "failed reading yaml from zip",
				})
				return
			}
			contentBytes, readErr := io.ReadAll(rc)
			rc.Close()
			if readErr != nil {
				JSONResponse(w, http.StatusBadRequest, map[string]string{
					"error": "failed reading yaml content",
				})
				return
			}

			yamlName = filepath.Base(name)
			yamlContent = string(contentBytes)
			continue
		}

		if !strings.HasPrefix(name, "assets/") {
			continue
		}

		rel := strings.TrimPrefix(name, "assets/")
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			continue
		}

		rc, openErr := f.Open()
		if openErr != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": "failed reading asset from zip",
			})
			return
		}
		contentBytes, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": "failed reading asset content",
			})
			return
		}
		assets[rel] = contentBytes
	}

	if yamlName == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "zip must contain one .yaml file",
		})
		return
	}

	validYamlName := regexp.MustCompile(`^[a-zA-Z0-9_-]+\.yaml$`)
	if !validYamlName.MatchString(yamlName) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml filename",
		})
		return
	}

	if len(assets) == 0 {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "zip must contain assets/ folder with files",
		})
		return
	}

	var tpl template.Template
	if err := yaml.Unmarshal([]byte(yamlContent), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml content",
		})
		return
	}

	if err := validateTemplatePayload(&tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if strings.TrimSpace(tpl.TemplateDir) == "" {
		dir, dirErr := buildTemplateDir(yamlName)
		if dirErr != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": dirErr.Error(),
			})
			return
		}
		tpl.TemplateDir = dir
	}

	validDir := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validDir.MatchString(tpl.TemplateDir) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid template_dir in yaml",
		})
		return
	}

	exists, err := h.repo.Exists(yamlName)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if exists {
		JSONResponse(w, http.StatusConflict, map[string]string{"error": "template already exists"})
		return
	}

	dirExists, err := h.repo.TemplateDirExists(tpl.TemplateDir)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if dirExists {
		JSONResponse(w, http.StatusConflict, map[string]string{"error": "assets directory already exists"})
		return
	}

	tpl.Info.System = false
	normalizedYaml, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := h.repo.Save(yamlName, string(normalizedYaml)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := h.repo.CreateTemplateDir(tpl.TemplateDir); err != nil {
		_ = h.repo.Delete(yamlName)
		JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	assetsBase := utils.GetBasePath(config.GetString("template_assets_dir"))
	dstRoot := filepath.Clean(filepath.Join(assetsBase, tpl.TemplateDir))
	dstRootPrefix := dstRoot + string(os.PathSeparator)

	rollback := func(msg string, status int) {
		_ = h.repo.Delete(yamlName)
		_ = h.repo.DeleteTemplateDir(tpl.TemplateDir)
		JSONResponse(w, status, map[string]string{"error": msg})
	}

	for rel, content := range assets {
		normalizedRel := filepath.ToSlash(rel)
		normalizedRel = strings.TrimPrefix(normalizedRel, "/")
		if strings.HasPrefix(normalizedRel, tpl.TemplateDir+"/") {
			normalizedRel = strings.TrimPrefix(normalizedRel, tpl.TemplateDir+"/")
		}
		cleanRel := filepath.Clean(normalizedRel)
		if cleanRel == "." || strings.HasPrefix(cleanRel, "..") {
			rollback("invalid asset path inside zip", http.StatusBadRequest)
			return
		}

		target := filepath.Clean(filepath.Join(dstRoot, cleanRel))
		if target != dstRoot && !strings.HasPrefix(target, dstRootPrefix) {
			rollback("invalid asset path inside zip", http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			rollback("failed creating asset directories", http.StatusInternalServerError)
			return
		}

		if err := os.WriteFile(target, content, 0644); err != nil {
			rollback("failed writing asset file", http.StatusInternalServerError)
			return
		}
	}

	JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"filename":     yamlName,
		"template_dir": tpl.TemplateDir,
		"imported":     true,
	})
}

func (h *TemplateHandler) ExportZip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := strings.TrimSpace(vars["filename"])
	if filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename is required",
		})
		return
	}

	if !strings.HasSuffix(strings.ToLower(filename), ".yaml") {
		filename += ".yaml"
	}

	meta, err := h.repo.LoadByFilename(filename)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "template not found",
		})
		return
	}

	assetsBase := utils.GetBasePath(config.GetString("template_assets_dir"))
	assetsDir := filepath.Clean(filepath.Join(assetsBase, meta.TemplateDir))

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", strings.TrimSuffix(filename, ".yaml")+".zip"))

	zw := zip.NewWriter(w)
	defer zw.Close()

	yamlEntry, err := zw.Create(filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create zip entry",
		})
		return
	}
	if _, err := yamlEntry.Write([]byte(meta.Content)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to write yaml to zip",
		})
		return
	}

	if info, statErr := os.Stat(assetsDir); statErr == nil && info.IsDir() {
		rootPrefix := assetsDir + string(os.PathSeparator)
		walkErr := filepath.Walk(assetsDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if path == assetsDir {
				return nil
			}

			rel, err := filepath.Rel(assetsDir, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if strings.HasPrefix(rel, "..") {
				return fmt.Errorf("invalid asset path")
			}

			entryPath := "assets/" + meta.TemplateDir + "/" + rel
			entryPath = filepath.ToSlash(entryPath)
			if strings.HasPrefix(filepath.Clean(path), rootPrefix) {
				if info.IsDir() {
					return nil
				}

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}
				header.Name = entryPath
				header.Method = zip.Deflate

				writer, err := zw.CreateHeader(header)
				if err != nil {
					return err
				}

				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				_, err = writer.Write(content)
				return err
			}

			return fmt.Errorf("invalid asset path")
		})

		if walkErr != nil {
			JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to package assets",
			})
			return
		}
	}
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	templates, err := h.repo.LoadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"templates": templates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TemplateHandler) GetByFilename(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	template, err := h.repo.GetTemplateByFilename(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) GetMetadataByFilename(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	template, err := h.repo.LoadByFilename(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {

	var req templateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" || req.Content == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename and content required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename must have .yaml extension",
		})
		return
	}

	name := strings.TrimSuffix(req.Filename, ".yaml")

	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if exists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template already exists",
		})
		return
	}

	dirExists, err := h.repo.TemplateDirExists(name)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if dirExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "assets directory already exists",
		})
		return
	}

	var tpl template.Template

	if err := yaml.Unmarshal([]byte(req.Content), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml content",
		})
		return
	}

	if err := validateTemplatePayload(&tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	tpl.Info.System = false
	tpl.TemplateDir, _ = buildTemplateDir(req.Filename)

	updatedYaml, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.Save(req.Filename, string(updatedYaml)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.CreateTemplateDir(tpl.TemplateDir); err != nil {
		h.repo.Delete(req.Filename)
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusCreated, tpl)
}

func (h *TemplateHandler) Clone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceFilename := vars["filename"]

	var req cloneTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if sourceFilename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "source filename required",
		})
		return
	}

	if req.NewFilename == "" || req.Name == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "new_filename and name required",
		})
		return
	}

	if !strings.HasSuffix(req.NewFilename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "new_filename must have .yaml extension",
		})
		return
	}

	newTemplateDir, err := buildTemplateDir(req.NewFilename)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	sourceExists, err := h.repo.Exists(sourceFilename)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if !sourceExists {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "source template not found",
		})
		return
	}

	newExists, err := h.repo.Exists(req.NewFilename)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if newExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template already exists",
		})
		return
	}

	dirExists, err := h.repo.TemplateDirExists(newTemplateDir)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if dirExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "assets directory already exists",
		})
		return
	}

	sourceMetadata, err := h.repo.LoadByFilename(sourceFilename)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "source template not found",
		})
		return
	}

	var tpl template.Template
	if err := yaml.Unmarshal([]byte(sourceMetadata.Content), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid source template yaml",
		})
		return
	}

	tpl.Info.System = false
	tpl.Info.Name = req.Name
	if req.Description != nil {
		tpl.Info.Description = *req.Description
	}
	tpl.TemplateDir = newTemplateDir

	updatedYAML, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.Save(req.NewFilename, string(updatedYAML)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.CopyTemplateDir(sourceMetadata.TemplateDir, tpl.TemplateDir); err != nil {
		_ = h.repo.Delete(req.NewFilename)
		_ = h.repo.DeleteTemplateDir(tpl.TemplateDir)
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to clone template assets",
		})
		return
	}

	JSONResponse(w, http.StatusCreated, tpl)
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {

	var req templateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" || req.Content == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename and content required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "template not found",
		})
		return
	}

	temp, _ := h.repo.GetTemplateByFilename(req.Filename)

	if temp.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	active, err := h.repoCam.HasActiveCampaignUsingTemplate(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to verify template usage",
		})
		return
	}

	if active {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template cannot be modified because there is an active campaign using it",
		})
		return
	}

	var tpl template.Template

	if err := yaml.Unmarshal([]byte(req.Content), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml content",
		})
		return
	}

	if err := validateTemplatePayload(&tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	tpl.Info.System = false

	expectedDir, err := buildTemplateDir(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	tpl.TemplateDir = expectedDir

	updatedYaml, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.Save(req.Filename, string(updatedYaml)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, tpl)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {

	var req deleteTemplateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "template not found",
		})
		return
	}

	count, err := h.repoCam.CountCampaignsUsingTemplateId(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to check template usage",
		})
		return
	}
	if count > 0 {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template is used by campaigns",
		})
		return
	}

	data, err := h.repo.LoadByFilename(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	data, _ = h.repo.LoadByFilename(req.Filename)

	templ, _ := h.repo.GetTemplateByFilename(req.Filename)

	if templ.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	if err := h.repo.Delete(req.Filename); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.DeleteTemplateDir(data.TemplateDir); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, data)
}
