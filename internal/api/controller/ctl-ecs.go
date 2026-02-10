package controller

import (
	"errors"
	"github.com/helays/ssh-proxy-plus/internal/api/service"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"net/http"

	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	vpc20160428 "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"helay.net/go/utils/v3/net/http/httpkit"
	"helay.net/go/utils/v3/net/http/request"
	"helay.net/go/utils/v3/net/http/response"
)

// CtlDescribeRegions 获取实例可用区
func (c *Controller) CtlDescribeRegions(w http.ResponseWriter, r *http.Request) {
	srv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	client, err := srv.InitECSClient()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	query := r.URL.Query()
	describeRegionsRequest := &ecs20140526.DescribeRegionsRequest{
		InstanceChargeType: tea.String(httpkit.QueryGet(query, "instance_charge_type", "PostPaid")),
		ResourceType:       tea.String(httpkit.QueryGet(query, "resource_type", "instance")),
		AcceptLanguage:     tea.String(httpkit.QueryGet(query, "accept_language", "zh-CN")),
	}

	runtime := &util.RuntimeOptions{}
	result, err := client.DescribeRegionsWithOptions(describeRegionsRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		response.SetReturnCode(w, r, *_err.StatusCode, *_err.Message, *_err.Data)
		return
	}
	response.SetReturnData(w, 0, result.Body.Regions)
}

// CtlDescribeAvailableResource 获取可用资源
func (c *Controller) CtlDescribeAvailableResource(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	regionId := httpkit.QueryGet(query, "region_id", "cn-hongkong")
	serv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	client, err := serv.InitECSClient(regionId)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	describeAvailableResourceRequest := &ecs20140526.DescribeAvailableResourceRequest{
		RegionId:            tea.String(regionId),
		InstanceChargeType:  tea.String(httpkit.QueryGet(query, "instance_charge_type", "PostPaid")),
		SpotStrategy:        tea.String("NoSpot"),
		DestinationResource: tea.String(httpkit.QueryGet(query, "destination_resource", "InstanceType")),
	}
	runtime := &util.RuntimeOptions{}
	result, err := client.DescribeAvailableResourceWithOptions(describeAvailableResourceRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		response.SetReturnCode(w, r, *_err.StatusCode, *_err.Message, *_err.Data)
		return
	}
	response.SetReturnData(w, 0, result.Body.AvailableZones)
}

// CtlDescribeVSwitches 查询交换机实例ID
func (c *Controller) CtlDescribeVSwitches(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	regionId := httpkit.QueryGet(query, "region_id", "cn-hongkong")
	serv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	aliClient, err := serv.InitAliVpcClient(regionId)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	describeVSwitchesRequest := &vpc20160428.DescribeVSwitchesRequest{
		RegionId: tea.String(regionId),
		PageSize: tea.Int32(50),
	}
	runtime := &util.RuntimeOptions{}
	result, err := aliClient.DescribeVSwitchesWithOptions(describeVSwitchesRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		response.SetReturnCode(w, r, *_err.StatusCode, *_err.Message, *_err.Data)
		return
	}
	response.SetReturnData(w, 0, result.Body.VSwitches)
}

// CtlDescribeSecurityGroups 查询安全组实例ID
func (c *Controller) CtlDescribeSecurityGroups(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	regionId := httpkit.QueryGet(query, "region_id", "cn-hongkong")
	serv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	aliClient, err := serv.InitECSClient(regionId)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	describeSecurityGroupsRequest := &ecs20140526.DescribeSecurityGroupsRequest{
		RegionId: tea.String(regionId),
		PageSize: tea.Int32(50),
	}
	runtime := &util.RuntimeOptions{}
	result, err := aliClient.DescribeSecurityGroupsWithOptions(describeSecurityGroupsRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		response.SetReturnCode(w, r, *_err.StatusCode, *_err.Message, *_err.Data)
		return
	}
	response.SetReturnData(w, 0, result.Body.SecurityGroups)
}

// CtlDescribeInstances 查询订单列表
func (c *Controller) CtlDescribeInstances(w http.ResponseWriter, r *http.Request) {
	var orders []model.AliEcsOrder
	tx := dal.GetDB().Omit("password").Order("id desc").Find(&orders)
	if tx.Error != nil {
		response.SetReturnErrorDisableLog(w, tx.Error, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, "成功", orders)
}

func (c *Controller) CtlCreateRunInstances(w http.ResponseWriter, r *http.Request) {
	postData, err := request.JsonDecode[model.AliEcsOrder](r)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	serv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	resp, err := serv.CreateInstance(&postData)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, resp)
}

func (c *Controller) CtlDelInstances(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		response.SetReturnErrorDisableLog(w, errors.New("id不能为空"), http.StatusForbidden)
		return
	}

	serv, err := service.NewECS()
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	resp, err := serv.DeleteInstance(id)
	if err != nil {
		response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
		return
	}
	response.SetReturnData(w, 0, resp)

}
