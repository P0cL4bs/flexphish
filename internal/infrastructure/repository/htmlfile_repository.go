package repository

import (
	"errors"
	"flexphish/internal/config"
	"flexphish/internal/domain/template"
	"os"
	"path/filepath"
	"strings"
)

type HtmlFileRepositoryObj struct {
	baseDir string
	trepo   template.TemplateRepository
}

func NewHtmlFileRepository(baseDir string, trepo template.TemplateRepository) template.HtmlfilesRepository {
	return &HtmlFileRepositoryObj{
		baseDir: baseDir,
		trepo:   trepo,
	}
}

func (r *HtmlFileRepositoryObj) resolvePath(templateDir, filename string) (string, error) {

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

func (r *HtmlFileRepositoryObj) List(templateDir string) ([]template.FileComplete, error) {

	dir := filepath.Join(r.baseDir, templateDir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []template.FileComplete

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}

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

func (r *HtmlFileRepositoryObj) Get(templateDir, filename string) (*template.FileComplete, error) {

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

func (r *HtmlFileRepositoryObj) Create(templateDir, filename, content string) error {

	path, err := r.resolvePath(templateDir, filename)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func (r *HtmlFileRepositoryObj) Update(templateDir, filename, content string) error {
	return r.Create(templateDir, filename, content)
}

func (r *HtmlFileRepositoryObj) Delete(templateDir, filename string) error {

	path, err := r.resolvePath(templateDir, filename)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

func (r *HtmlFileRepositoryObj) Exists(templateDir, filename string) (bool, error) {

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

func (r *HtmlFileRepositoryObj) GetAllByTemplateFilename(templateFilename string) ([]template.FileComplete, error) {

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
	)

	pattern := filepath.Join(basePath, "*.html")

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

func (r *HtmlFileRepositoryObj) CreateByTemplateFilename(templateFilename, filename, content string) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := meta.TemplateDir

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

func (r *HtmlFileRepositoryObj) UpdateByTemplateFilename(templateFilename, filename, content string) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := meta.TemplateDir

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

func (r *HtmlFileRepositoryObj) DeleteByTemplateFilename(templateFilename, filename string) (*template.FileComplete, error) {

	meta, err := r.trepo.LoadByFilename(templateFilename)
	if err != nil {
		return nil, errors.New("template not found")
	}

	templateDir := meta.TemplateDir

	file, err := r.Get(templateDir, filename)
	if err != nil {
		return nil, errors.New("file not found")
	}

	if err := r.Delete(templateDir, filename); err != nil {
		return nil, err
	}

	return file, nil
}
