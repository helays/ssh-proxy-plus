package service

import (
	"errors"
	"fmt"
	cmp_proxy "github.com/helays/ssh-proxy-plus/internal/component/cmp-proxy"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	dal_connect "github.com/helays/ssh-proxy-plus/internal/dal/dal-connect"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"github.com/helays/ssh-proxy-plus/internal/types"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"
	"gorm.io/gorm"
	"helay.net/go/utils/v3/logger/ulogs"
	"helay.net/go/utils/v3/tools"
)

type ForWardService struct {
}

func NewForWardService() *ForWardService {
	return &ForWardService{}
}

func (s *ForWardService) filterAndSaveRequestData(r *http.Request) (*model.Connect, error) {
	var c = model.Connect{}
	sType, err := strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		return nil, fmt.Errorf("参数 type 格式错误 %v", err)
	}
	c.Stype = types.SSHValidType(sType)
	c.Id = strings.TrimSpace(r.PostFormValue("id"))
	c.Pid = strings.TrimSpace(r.PostFormValue("pid"))
	isNew := false
	if c.Id == "" {
		c.Id = bson.NewObjectId().Hex()
		isNew = true
	}
	c.Lname = strings.TrimSpace(r.PostFormValue("lname"))
	c.Saddr = strings.TrimSpace(r.PostFormValue("saddr"))
	c.User = strings.TrimSpace(r.PostFormValue("user"))
	tmpConnect := strings.TrimSpace(r.PostFormValue("connect"))
	c.Connect = types.ForwardType(tmpConnect)
	c.Remote = strings.TrimSpace(r.PostFormValue("remote"))
	c.Listen = strings.TrimSpace(r.PostFormValue("listen"))
	var his model.Connect
	db := dal.GetDB()
	if !isNew {
		if err = db.Model(model.Connect{}).Where("id like ?", c.Id).Take(&his).Error; err != nil {
			return nil, fmt.Errorf("数据库查询错误 %v", err)
		}
		c.Active = his.Active
	}
	c.Passwd = strings.TrimSpace(r.PostFormValue("passwd"))
	if c.Passwd == "" {
		c.Passwd = his.Passwd
	}
	if sType != 1 && sType != 2 {
		return nil, fmt.Errorf("密码类型错误")

	}
	return &c, dal_connect.SaveData(&c, isNew)
}

func (s *ForWardService) Create(r *http.Request) error {
	c, err := s.filterAndSaveRequestData(r)
	if err != nil {
		return err
	}
	return RunForwardInstance(c.Id)
}

func (s *ForWardService) Update(r *http.Request) error {
	c, err := s.filterAndSaveRequestData(r)
	if err != nil {
		return err
	}
	// 发送停止信号
	ulogs.Infof("停止隧道 %s", c.Id)
	cmp_proxy.Stop(c.Id)
	return RunForwardInstance(c.Id)
}

func RunForwardInstance(cId string) error {
	dal_connect.ReadConnect2Cache() // 同步所有连接信息到内存
	result := cmp_proxy.FindConnectByID(cId)
	if result != nil {
		ulogs.Infof("启动隧道 %s", cId)
		cmp_proxy.Start(*result)
		return nil
	}
	return fmt.Errorf("内存中未找到该配置，请核对信息。")
}

func (s *ForWardService) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("参数 id 不能为空")
	}
	// 停止该连接
	ulogs.Infof("停止隧道 %s", id)
	cmp_proxy.Stop(id)
	db := dal.GetDB()
	if err := db.Where("id like ?", id).Delete(&model.Connect{}).Error; err != nil {
		return fmt.Errorf("数据库删除错误 %v", err)
	}
	dal_connect.ReadConnect2Cache() // 同步所有连接信息到内存
	return nil
}

func (s *ForWardService) Stop(r *http.Request) error {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		return fmt.Errorf("参数 id 不能为空")
	}
	var data model.Connect
	db := dal.GetDB()
	if err := db.Model(model.Connect{}).Where("id like ?", id).Take(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("数据库未找到该配置")
		}
		return fmt.Errorf("数据库查询错误 %v", err)
	}
	toActive := tools.Ternary(data.Active == "Y", "N", "Y")
	tx := db.Model(model.Connect{}).Where("id like ?", id).Update("active", toActive)
	if err := tx.Error; err != nil {
		return fmt.Errorf("隧道更新连接状态错误 %v", err)
	}
	dal_connect.ReadConnect2Cache() // 更新缓存信息
	if toActive == "N" {
		cmp_proxy.Stop(id)
		return nil
	}
	result := cmp_proxy.FindConnectByID(id)
	if result != nil {
		cmp_proxy.Start(*result)
		return nil
	}
	return fmt.Errorf("内存中未找到该配置，请核对信息。")
}
