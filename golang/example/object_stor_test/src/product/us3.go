package product

import (
	ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
	"src/common"
)

type Us3 struct {
	path, objName string
}

func NewUs3(path, objName string) *Us3 {
	return &Us3{path: path, objName: objName}
}

func (u *Us3) Upload() error {
	req, err := ufsdk.NewFileRequest(common.Config, nil)
	if err != nil {
		return err
	}

	err = req.AsyncUpload(u.path, u.objName, "", 10)
	if err != nil {
		return err
	}

	return nil
}
