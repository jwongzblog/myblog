package product

type Oss struct {
	path, objName string
}

func NewOss(path, objName string) *Oss {
	return &Oss{path: path, objName: objName}
}

func (u *Oss) Upload() error {
	return nil
}
