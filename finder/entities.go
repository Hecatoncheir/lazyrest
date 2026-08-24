package finder

type Directory struct {
	Name        string
	Path        string
	Directories []Directory
	Files       []File
	Warnings    []string
}

type File struct {
	Name string
	Path string
}
