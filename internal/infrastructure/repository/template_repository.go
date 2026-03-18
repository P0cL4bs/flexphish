package repository

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"flexphish/internal/domain/template"
	"flexphish/pkg/utils"

	"gopkg.in/yaml.v2"
)

var validFilename = regexp.MustCompile(`^[a-zA-Z0-9_-]+\.yaml$`)

func validateFilename(filename string) error {
	if !validFilename.MatchString(filename) {
		return errors.New("invalid filename")
	}
	return nil
}

type TemplateRepositoryObj struct {
	templateDir       string
	templateAssetsDir string
}

func (r *TemplateRepositoryObj) resolvePath(filename string) (string, error) {

	if err := validateFilename(filename); err != nil {
		return "", err
	}

	base := filepath.Clean(r.templateDir)
	full := filepath.Join(base, filename)

	if !strings.HasPrefix(full, base) {
		return "", errors.New("invalid path")
	}

	return full, nil
}

func (r *TemplateRepositoryObj) resolveTemplateAssetsDir(templateDir string) (string, error) {

	cleanBase := filepath.Clean(r.templateAssetsDir)

	full := filepath.Join(cleanBase, templateDir)
	full = filepath.Clean(full)

	baseWithSep := cleanBase + string(os.PathSeparator)

	if full != cleanBase && !strings.HasPrefix(full, baseWithSep) {
		return "", errors.New("invalid template_dir")
	}

	return full, nil
}

func NewTemplateRepository(cfg map[string]interface{}) template.TemplateRepository {
	r := &TemplateRepositoryObj{}

	if dir, ok := cfg["template_dir"].(string); ok {
		r.templateDir = utils.GetBasePath(dir)
	}

	if dirAssets, ok := cfg["template_assets_dir"].(string); ok {
		r.templateAssetsDir = utils.GetBasePath(dirAssets)
	}

	return r
}

func (r *TemplateRepositoryObj) LoadAll() (map[string]*template.TemplateMetadata, error) {
	files, err := filepath.Glob(filepath.Join(r.templateDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	templates := make(map[string]*template.TemplateMetadata)

	for _, file := range files {
		meta, err := r.loadTemplate(file)
		if err != nil {
			continue
		}

		templates[filepath.Base(file)] = meta
	}

	return templates, nil
}

func (r *TemplateRepositoryObj) loadTemplate(path string) (*template.TemplateMetadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		template.Info
		TemplateDir string `yaml:"template_dir"`
	}

	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	assetBasePath, err := r.resolveTemplateAssetsDir(parsed.TemplateDir)
	if err != nil {
		return nil, err
	}

	htmlFiles := []template.HtmlFile{}
	htmlPaths, _ := filepath.Glob(filepath.Join(assetBasePath, "*.html"))

	for _, html := range htmlPaths {
		htmlStat, err := os.Stat(html)
		if err != nil {
			continue
		}

		htmlFiles = append(htmlFiles, template.HtmlFile{
			Filename: htmlStat.Name(),
			Path:     html,
			Size:     htmlStat.Size(),
			ModTime:  htmlStat.ModTime(),
		})
	}

	return &template.TemplateMetadata{
		Content:     string(content),
		Filename:    stat.Name(),
		Name:        parsed.Info.Name,
		Author:      parsed.Info.Author,
		Description: parsed.Info.Description,
		Category:    parsed.Info.Category,
		Info:        parsed.Info,
		Tags:        parsed.Info.Tags,
		TemplateDir: parsed.TemplateDir,
		Size:        stat.Size(),
		ModTime:     stat.ModTime(),
		IsDir:       stat.IsDir(),
		Mode:        stat.Mode().String(),
		HtmlFiles:   htmlFiles,
	}, nil
}

func (r *TemplateRepositoryObj) LoadByFilename(filename string) (*template.TemplateMetadata, error) {
	path := filepath.Join(r.templateDir, filename)
	return r.loadTemplate(path)
}

func (r *TemplateRepositoryObj) loadTemplateFromPath(path string) (*template.Template, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tmpl template.Template

	if err := yaml.Unmarshal(content, &tmpl); err != nil {
		return nil, err
	}
	tmpl.TemplateDir = filepath.Join(r.templateAssetsDir, tmpl.TemplateDir)

	return &tmpl, nil
}

func (r *TemplateRepositoryObj) GetTemplateByFilename(t_filename string) (*template.Template, error) {
	return r.loadTemplateFromPath(filepath.Join(r.templateDir, t_filename))
}

func (r *TemplateRepositoryObj) Exists(filename string) (bool, error) {

	path, err := r.resolvePath(filename)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (r *TemplateRepositoryObj) TemplateDirExists(templateDir string) (bool, error) {

	path, err := r.resolveTemplateAssetsDir(templateDir)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (r *TemplateRepositoryObj) CreateTemplateDir(templateDir string) error {

	path, err := r.resolveTemplateAssetsDir(templateDir)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0755)
}

func (r *TemplateRepositoryObj) Save(filename, content string) error {

	path, err := r.resolvePath(filename)
	if err != nil {
		return err
	}

	var tpl template.Template

	if err := yaml.Unmarshal([]byte(content), &tpl); err != nil {
		return err
	}

	return utils.SaveYAMLFile(path, tpl)
}

func (r *TemplateRepositoryObj) Delete(filename string) error {

	path, err := r.resolvePath(filename)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("template not found")
		}
		return err
	}

	return nil
}

func (r *TemplateRepositoryObj) DeleteTemplateDir(templateDir string) error {

	path, err := r.resolveTemplateAssetsDir(templateDir)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}
