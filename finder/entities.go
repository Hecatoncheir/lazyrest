package finder

type Directory struct {
	Name        string
	Path        string
	Directories []Directory
	Files       []File
}

type File struct {
	Name string
	Path string
}
