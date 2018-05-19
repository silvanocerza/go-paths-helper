package paths

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Path represents a path
type Path struct {
	path               string
	cachedFileInfo     os.FileInfo
	cachedFileInfoTime time.Time
}

// New creates a new Path object. If path is the empty string
// then nil is returned.
func New(path string) *Path {
	if path == "" {
		return nil
	}
	return &Path{path: path}
}

func (p *Path) setCachedFileInfo(info os.FileInfo) {
	p.cachedFileInfo = info
	p.cachedFileInfoTime = time.Now()
}

// Stat returns a FileInfo describing the named file. The result is
// cached internally for next queries. To ensure that the cached
// FileInfo entry is updated just call Stat again.
func (p *Path) Stat() (os.FileInfo, error) {
	info, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}
	p.setCachedFileInfo(info)
	return info, nil
}

func (p *Path) stat() (os.FileInfo, error) {
	if p.cachedFileInfo != nil {
		if p.cachedFileInfoTime.Add(50 * time.Millisecond).After(time.Now()) {
			return p.cachedFileInfo, nil
		}
	}
	return p.Stat()
}

// Clone create a copy of the Path object
func (p *Path) Clone() *Path {
	return New(p.path)
}

// Join create a new Path by joining the provided paths
func (p *Path) Join(paths ...string) *Path {
	return New(filepath.Join(p.path, filepath.Join(paths...)))
}

// JoinPath create a new Path by joining the provided paths
func (p *Path) JoinPath(paths ...*Path) *Path {
	res := p.Clone()
	for _, path := range paths {
		res = res.Join(path.path)
	}
	return res
}

// Base Base returns the last element of path
func (p *Path) Base() string {
	return filepath.Base(p.path)
}

// RelTo returns a relative Path that is lexically equivalent to r when
// joined to the current Path
func (p *Path) RelTo(r *Path) (*Path, error) {
	rel, err := filepath.Rel(p.path, r.path)
	if err != nil {
		return nil, err
	}
	return New(rel), nil
}

// Abs returns the absolute path of the current Path
func (p *Path) Abs() (*Path, error) {
	abs, err := filepath.Abs(p.path)
	if err != nil {
		return nil, err
	}
	return New(abs), nil
}

// IsAbs returns true if the Path is absolute
func (p *Path) IsAbs() bool {
	return filepath.IsAbs(p.path)
}

// ToAbs transofrm the current Path to the corresponding absolute path
func (p *Path) ToAbs() error {
	abs, err := filepath.Abs(p.path)
	if err != nil {
		return err
	}
	p.path = abs
	return nil
}

// Clean Clean returns the shortest path name equivalent to path by
// purely lexical processing
func (p *Path) Clean() *Path {
	return New(filepath.Clean(p.path))
}

// Parent returns all but the last element of path, typically the path's
// directory or the parent directory if the path is already a directory
func (p *Path) Parent() *Path {
	return New(filepath.Dir(p.path))
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error
func (p *Path) MkdirAll() error {
	return os.MkdirAll(p.path, os.FileMode(0755))
}

// Remove removes the named file or directory
func (p *Path) Remove() error {
	return os.Remove(p.path)
}

// FollowSymLink transforms the current path to the path pointed by the
// symlink if path is a symlink, otherwise it does nothing
func (p *Path) FollowSymLink() error {
	resolvedPath, err := filepath.EvalSymlinks(p.path)
	if err != nil {
		return err
	}
	p.path = resolvedPath
	p.cachedFileInfo = nil
	return nil
}

// Exist return true if the path exists
func (p *Path) Exist() (bool, error) {
	_, err := p.stat()
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsDir return true if the path exists and is a directory
func (p *Path) IsDir() (bool, error) {
	info, err := p.stat()
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadDir returns a PathList containing the content of the directory
// pointed by the current Path
func (p *Path) ReadDir() (PathList, error) {
	infos, err := ioutil.ReadDir(p.path)
	if err != nil {
		return nil, err
	}
	paths := PathList{}
	for _, info := range infos {
		path := p.Clone().Join(info.Name())
		path.setCachedFileInfo(info)
		paths.Add(path)
	}
	return paths, nil
}

// CopyTo copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func (p *Path) CopyTo(dst *Path) error {
	in, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst.path)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	si, err := p.Stat()
	if err != nil {
		return err
	}

	err = os.Chmod(dst.path, si.Mode())
	if err != nil {
		return err
	}

	return nil
}

// Chtimes changes the access and modification times of the named file,
// similar to the Unix utime() or utimes() functions.
func (p *Path) Chtimes(atime, mtime time.Time) error {
	return os.Chtimes(p.path, atime, mtime)
}

// ReadFile reads the file named by filename and returns the contents
func (p *Path) ReadFile() ([]byte, error) {
	return ioutil.ReadFile(p.path)
}

// WriteFile writes data to a file named by filename. If the file
// does not exist, WriteFile creates it otherwise WriteFile truncates
// it before writing.
func (p *Path) WriteFile(data []byte) error {
	return ioutil.WriteFile(p.path, data, os.FileMode(0644))
}

func (p *Path) String() string {
	return p.path
}