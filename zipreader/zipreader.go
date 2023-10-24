package zipreader

import (
	"fmt"
	"strings"
	"archive/zip"
	"errors"
	"io"
)

type MyReader struct {
	reader *zip.ReadCloser
	filtered_names map[string]string
	filter_s string
	name string
}

func (m *MyReader) Close(){
	if m.reader!=nil {
		m.reader.Close()
	}
}

func (m *MyReader) FileNames() []string{
	ret := make([]string, 0)
	for v := range m.filtered_names {
		ret = append(ret, v)
	}
	return ret
}

func (m *MyReader) ReadFile(name string) (int64, []byte, error){
	if name=="" {
		return 0, nil, errors.New("name is empty")
	}
	if _, has := m.filtered_names[name]; !has {
		return 0, nil, errors.New(fmt.Sprintf("file '%s' not found in archive '%s'", name, m.name))
	}
	file, err := m.reader.Open(name)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()

	ret := make([]byte, size)
	size1, err1 := io.ReadFull(file, ret)
	if err1 != nil && err1!=io.EOF {
		return 0, nil, err1
	}
	if int64(size1)!=size {
		return 0, nil, errors.New(fmt.Sprintf("was read wrong size of data of file '%s' in archive '%s'. expect %d, got %d", name, m.name, size, size1))
	}
	return size, ret, nil
}

func (m *MyReader) RawFilter() string{
	return m.filter_s
}

func (m *MyReader) Name() string{
	return m.name
}

func CreateReader(name, filter string) (*MyReader, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, 0)
	for _, f := range r.File {
		if filter=="" || strings.HasSuffix(f.Name, filter) {
			names[f.Name] = f.Name
		}
	}
	return &MyReader{r, names, filter, name}, nil
}

func (m *MyReader) SizeOfFile(name string) string {
	if name=="" {
		return ""
	}
	if _, has := m.filtered_names[name]; !has {
		return ""
	}
	file, err := m.reader.Open(name)
	if err != nil {
		return ""
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	return bytesFormatted(size)
}

func bytesFormatted(size int64) string {
	if size<0 {
		return ""
	}
	if size<1024 {
		return fmt.Sprintf("%dB", size)
	}
	formats := [8]string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "YB"}
	i:=0
	size_f := float64(size)
	for size_f>1024 && i<len(formats) {
		i++;
		size_f = size_f/1024.0
	}
	return fmt.Sprintf("%0.1f%s", size_f, formats[i])
}