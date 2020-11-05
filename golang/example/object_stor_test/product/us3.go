package product

import (
	"../common"
	ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
)

type Us3 struct {
	path, objName string
}

func NewUs3(path, objName string) *Us3 {
	return &Us3{path: path, objName: objName}
}

func (u *Us3) Upload() error {
	config, err := ufsdk.LoadConfig(common.Us3ConfigFile)
	if err != nil {
		return err
	}

	req, err := ufsdk.NewFileRequest(config, nil)
	if err != nil {
		return err
	}

	err = req.AsyncUpload(u.path, u.objName, "", 10)
	if err != nil {
		return err
	}

	return nil
}
