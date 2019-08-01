package api

import (
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/api/as"
	"github.com/brocaar/lora-app-server/internal/api/external"
	"github.com/brocaar/lora-app-server/internal/api/js"
	"github.com/brocaar/lora-app-server/internal/config"
)

func Setup(conf config.Config) error {
	//应用服务器api，不同应用的api是隔离的
	if err := as.Setup(conf); err != nil {
		return errors.Wrap(err, "setup application-server api error")
	}

	//指lora app server的api接口
	if err := external.Setup(conf); err != nil {
		return errors.Wrap(err, "setup external api error")
	}

	//js专门的api接口
	if err := js.Setup(conf); err != nil {
		return errors.Wrap(err, "setup join-server api error")
	}

	return nil
}
