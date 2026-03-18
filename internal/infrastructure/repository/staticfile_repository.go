package repository

import (
	"errors"
	"flexphish/internal/config"
	"flexphish/internal/domain/template"
	"os"
	"path/filepath"
	"strings"
)

type StaticFileRepositoryObj struct {
	baseDir string
	trepo   template.TemplateRepository
}

func NewStaticFileRepository(baseDir string, trepo template.TemplateRepository) template.StaticFileRepository {
	return &StaticFileRepositoryObj{
		baseDir: baseDir,
		trepo:   trepo,
	}
}

func (r *StaticFileRepositoryObj) resolvePath(templateDir, filename string) (string, error) {

	fullDir := filepath.Join(r.baseDir, templateDir)
	fullPath := filepath.Join(fullDir, filename)

	absBase, err := filepath.Abs(r.baseDir)
	if err != nil {
		return "", err
	}

	absFile, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absBase, absFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", os.ErrPermission
	}

	return absFile, nil
}

func (r *StaticFileRepositoryObj) List(templateDir string) ([]template.FileComplete, error) {

	dir := filepath.Join(r.baseDir, templateDir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []template.FileComplete

	for _, e := range entries {

		fullPath := filepath.Join(dir, e.Name())

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		files = append(files, template.FileComplete{
			Filename: e.Name(),
			Path:     fullPath,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Content:  string(content),
		})
	}

	return files, nil
}

func (r *StaticFileRepositoryObj) Get(templateDir, filename string) (*template.FileComplete, error) {

	path, err := r.resolvePath(templateDir, filename)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &template.FileComplete{
		Filename: filename,
		Path:     path,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		Content:  string(content),
	}, nil
}

func (r *StaticFileRepositoryObj) Create(templateDir, filename string, content []byte) error {

	path, err := r.resolvePath(templateDir, filename)
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}

func (r *StaticFileRepositoryObj) Update(templateDir, filename, content string) error {
	return r.Create(templateDir, filename, []byte(content))
}

func (r *StaticFileRepositoryObj) Delete(templateDir, filename string) error {

	path, err := r.resolvePath(templateDir, filename)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

func (r *StaticFileRepositoryObj) Exists(templateDir, filename string) (bool, error) {

	path, err := r.resolvePath(templateDir, filename)
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

func (r *StaticFileRepositoryObj) GetAllByTemplateFilename(templateFilename string) ([]template.FileComplete, error) {

	exists, err := r.trepo.Exists(templateFilename)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("template does not exist")
	}

	data, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, err
	}

	basePath := filepath.Join(
		config.GetString("template_assets_dir"),
		data.TemplateDir,
		"static",
	)

	pattern := filepath.Join(basePath, "*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return []template.FileComplete{}, nil
	}

	var files []template.FileComplete

	for _, path := range matches {

		stat, err := os.Stat(path)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		files = append(files, template.FileComplete{
			Filename: stat.Name(),
			Path:     path,
			Size:     stat.Size(),
			ModTime:  stat.ModTime(),
			Content:  string(content),
		})
	}

	return files, nil
}

func (r *StaticFileRepositoryObj) CreateByTemplateFilename(templateFilename, filename string, content []byte) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := filepath.Join(meta.TemplateDir, "static")

	if err := os.MkdirAll(filepath.Join(r.baseDir, templateDir), 0755); err != nil {
		return nil, err
	}

	exists, err := r.Exists(templateDir, filename)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("file already exists")
	}

	if err := r.Create(templateDir, filename, content); err != nil {
		return nil, err
	}

	return r.Get(templateDir, filename)
}

func (r *StaticFileRepositoryObj) UpdateByTemplateFilename(templateFilename, filename, content string) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := filepath.Join(meta.TemplateDir, "static")

	exists, err := r.Exists(templateDir, filename)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("file not found")
	}

	if err := r.Update(templateDir, filename, content); err != nil {
		return nil, err
	}

	return r.Get(templateDir, filename)
}

func (r *StaticFileRepositoryObj) DeleteByTemplateFilename(templateFilename, filename string) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := filepath.Join(meta.TemplateDir, "static")

	file, err := r.Get(templateDir, filename)
	if err != nil {
		return nil, errors.New("file not found")
	}

	if err := r.Delete(templateDir, filename); err != nil {
		return nil, err
	}

	return file, nil
}
