package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/helays/ssh-proxy-plus/configs"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/dal"
	dal_connect "github.com/helays/ssh-proxy-plus/internal/dal/dal-connect"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	vpc20160428 "github.com/alibabacloud-go/vpc-20160428/v6/client"
	"github.com/helays/utils/v2/logger/ulogs"
	"gopkg.in/mgo.v2/bson"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EcsService struct {
	cfg             *configs.Config
	accessKeyId     string
	accessKeySecret string
}

func NewECS() (*EcsService, error) {
	srv := EcsService{
		cfg: configs.Get(),
	}
	if accessKeyId, ok := cache.SysConfig.Load("access_key_id"); !ok || accessKeyId.Value == "" {
		return nil, fmt.Errorf("请先配置阿里云密钥 Access Key ID")
	} else {
		srv.accessKeyId = accessKeyId.Value
	}
	if accessKeySecret, ok := cache.SysConfig.Load("access_key_secret"); !ok || accessKeySecret.Value == "" {
		return nil, fmt.Errorf("请先配置阿里云密钥 Access Key Secret")
	} else {
		srv.accessKeySecret = accessKeySecret.Value
	}

	return &srv, nil
}

// InitECSClient 创建ECS客户端
func (s *EcsService) InitECSClient(endPoint ...string) (_result *ecs20140526.Client, _err error) {
	cg := &openapi.Config{
		AccessKeyId:     tea.String(s.accessKeyId),
		AccessKeySecret: tea.String(s.accessKeySecret),
	}
	if len(endPoint) > 0 && endPoint[0] != "" {
		cg.Endpoint = tea.String("ecs." + endPoint[0] + ".aliyuncs.com")
	} else {
		cg.Endpoint = tea.String("ecs.cn-shanghai.aliyuncs.com")
	}
	return ecs20140526.NewClient(cg)
}

// InitAliVpcClient 初始化VPC客户端
func (s *EcsService) InitAliVpcClient(endPoint ...string) (_result *vpc20160428.Client, _err error) {
	cg := &openapi.Config{
		AccessKeyId:     tea.String(s.accessKeyId),
		AccessKeySecret: tea.String(s.accessKeySecret),
	}
	if len(endPoint) > 0 && endPoint[0] != "" {
		cg.Endpoint = tea.String("vpc." + endPoint[0] + ".aliyuncs.com")
	} else {
		cg.Endpoint = tea.String("vpc.cn-shanghai.aliyuncs.com")
	}

	return vpc20160428.NewClient(cg)
}

func (s *EcsService) DeleteInstance(id string) (any, error) {
	var order model.AliEcsOrder
	db := dal.GetDB()
	if err := db.Take(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败 %v", err)
	}
	if order.ConnectId != "" {
		serv := NewForWardService()
		if err := serv.Delete(order.ConnectId); err != nil {
			return nil, fmt.Errorf("删除隧道实例失败 %v", err)
		}
	}

	var (
		err       error
		requestId string
	)
	if order.InstanceId != "" {
		requestId, err = s.deleteInstance(&order)
		if err != nil {
			return nil, err
		}
	}
	if err = db.Delete(&order).Error; err != nil {
		return requestId, fmt.Errorf("删除工单数据失败，请求号%s，错误信息：%v", requestId, err)
	}
	return requestId, nil
}

func (s *EcsService) deleteInstance(order *model.AliEcsOrder) (string, error) {
	client, err := s.InitECSClient(order.RegionId)
	if err != nil {
		return "", err
	}
	deleteInstanceRequest := &ecs20140526.DeleteInstanceRequest{
		InstanceId:            tea.String(order.InstanceId),
		Force:                 tea.Bool(true),  // 是否强制释放运行中（Running）的实例
		TerminateSubscription: tea.Bool(false), // 是否释放已到期的包年包月实例
	}
	runtime := &util.RuntimeOptions{}
	result, err := client.DeleteInstanceWithOptions(deleteInstanceRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		return "", fmt.Errorf("删除实例失败，状态码：%d，错误信息：%s，原始信息： %s",
			tea.IntValue(_err.StatusCode),
			tea.StringValue(_err.Message),
			tea.StringValue(_err.Data))
	}
	return tea.StringValue(result.Body.RequestId), nil
}

func (s *EcsService) CreateInstance(postData *model.AliEcsOrder) (any, error) {
	if postData.LocalListenAddr == "" {
		return nil, fmt.Errorf("请填写本地监听地址")
	}
	client, err := s.InitECSClient(postData.RegionId)
	if err != nil {
		return nil, err
	}
	requestInfo := &ecs20140526.RunInstancesRequest{
		RegionId:                &postData.RegionId,
		ImageId:                 &postData.ImageId,
		InstanceType:            &postData.InstanceType,
		PasswordInherit:         &postData.PasswordInherit,
		AutoRenew:               &postData.AutoRenew,
		InstanceChargeType:      &postData.InstanceChargeType,
		AutoPay:                 &postData.AutoPay,
		InternetChargeType:      &postData.InternetChargeType,
		InternetMaxBandwidthIn:  &postData.InternetMaxBandwidth,
		InternetMaxBandwidthOut: &postData.InternetMaxBandwidth,
		DryRun:                  &postData.DryRun,
		SecurityGroupId:         &postData.SecurityGroupId,
		Password:                &postData.Password,
		SystemDisk: &ecs20140526.RunInstancesRequestSystemDisk{
			Category: &postData.SystemDisk.Category,
			Size:     tea.String(strconv.Itoa(int(postData.SystemDisk.Size))),
		},
		IoOptimized:                 &postData.IoOptimized,
		SecurityEnhancementStrategy: &postData.SecurityEnhancementStrategy,
		VSwitchId:                   &postData.VSwitchId,
	}
	if postData.InstanceChargeType == "PostPaid" && postData.AutoReleaseTime > 0 {
		//当前时间加上 postData.AutoReleaseTime小时，最终输出yyyy-MM-ddTHH:mm:ssZ这样的时间格式
		// 获取当前时间并转换为UTC时区
		nowUTC := time.Now().UTC()

		// 计算加上N小时后的时间
		futureTimeUTC := nowUTC.Add(time.Hour * time.Duration(postData.AutoReleaseTime))

		// 使用ISO 8601格式化时间，Z表示零时区（即UTC+0）
		formattedTime := futureTimeUTC.Format("2006-01-02T15:04:05Z")
		requestInfo.AutoReleaseTime = &formattedTime
	}
	db := dal.GetDB()
	if err = db.Create(&postData).Error; err != nil {
		return nil, fmt.Errorf("订单保存数据库失败")
	}

	runtime := &util.RuntimeOptions{}
	result, err := client.RunInstancesWithOptions(requestInfo, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		var errData map[string]any
		_ = json.NewDecoder(bytes.NewBufferString(tea.StringValue(_err.Data))).Decode(&errData)
		err = db.Model(model.AliEcsOrder{}).Where("id=?", postData.Id).Updates(map[string]any{
			"order_status": tea.IntValue(_err.StatusCode),
			"err_message":  tea.StringValue(_err.Message),
			"err_data":     datatypes.JSONMap(errData),
		}).Error
		if err != nil {
			return nil, fmt.Errorf("创建ECS实例失败，同时保存错误信息失败，状态码：%d，错误信息：%s，原始数据：%s，数据库错误信息：%v",
				tea.IntValue(_err.StatusCode),
				tea.StringValue(_err.Message),
				tea.StringValue(_err.Data),
				err,
			)
		}
		return nil, fmt.Errorf("创建ECS实例失败，状态码：%d，错误信息：%s，原始数据：%s",
			tea.IntValue(_err.StatusCode),
			tea.StringValue(_err.Message),
			tea.StringValue(_err.Data),
		)
	}

	postData.InstanceId = tea.StringValue(result.Body.InstanceIdSets.InstanceIdSet[0])

	tx := db.Model(model.AliEcsOrder{}).Where("id = ?", postData.Id).Updates(map[string]any{
		"order_status": tea.Int32Value(result.StatusCode),
		"request_id":   tea.StringValue(result.Body.RequestId),
		"order_id":     tea.StringValue(result.Body.OrderId),
		"trade_price":  tea.Float32Value(result.Body.TradePrice),
		"instance_id":  postData.InstanceId,
	})
	if err = tx.Error; err != nil {
		ulogs.Errorf("更新订单状态失败 %v", err)
	}

	go func() {
		time.Sleep(5 * time.Second) // 等待5秒后，再判断订单状态是否开通
		status404 := 0
		tck := time.NewTicker(10 * time.Second)
		defer tck.Stop()
		for range tck.C {
			if !s.describeInstanceAttribute(client, postData) {
				if postData.QueryStatus == 404 && status404 < 5 {
					status404++
				} else {
					break
				}
			}
		}
	}()

	return nil, nil
}

func (s *EcsService) describeInstanceAttribute(client *ecs20140526.Client, order *model.AliEcsOrder) bool {
	describeInstanceAttributeRequest := &ecs20140526.DescribeInstanceAttributeRequest{
		InstanceId: tea.String(order.InstanceId),
	}
	runtime := &util.RuntimeOptions{}
	db := dal.GetDB()
	result, err := client.DescribeInstanceAttributeWithOptions(describeInstanceAttributeRequest, runtime)
	if err != nil {
		var _err = &tea.SDKError{}
		var _t *tea.SDKError
		if errors.As(err, &_t) {
			_err = _t
		}
		var errData map[string]any
		_ = json.NewDecoder(bytes.NewBufferString(tea.StringValue(_err.Data))).Decode(&errData)
		err = db.Model(model.AliEcsOrder{}).Where("id=?", order.Id).Updates(map[string]any{
			"query_status":      tea.IntValue(_err.StatusCode),
			"query_err_message": tea.StringValue(_err.Message),
			"query_err_data":    datatypes.JSONMap(errData),
		}).Error
		ulogs.Checkerr(err, "查询实例开通状态失败", "响应信息入库失败", order.Id)
		return false
	}
	order.RunStatus = tea.StringValue(result.Body.Status)
	order.PublicIpAddress = tea.StringValue(result.Body.PublicIpAddress.IpAddress[0])
	order.QueryStatus = int(tea.Int32Value(result.StatusCode))
	tx := db.Model(model.AliEcsOrder{}).Where("id=?", order.Id).Updates(map[string]any{
		"query_status":      order.QueryStatus,
		"run_status":        order.RunStatus,
		"public_ip_address": order.PublicIpAddress,
	})
	if err = tx.Error; err != nil {
		ulogs.Errorf("更新订单状态失败 %v", err)
	}
	switch order.RunStatus {
	case "Pending": // 创建中
		return true
	case "Running": // 运行中
		// 这里还需要将ip信息拿出来，去自动添加动态代理
		time.Sleep(3 * time.Second)
		s.runProxy(order)
		return false
	case "Starting": // 启动中
		return true
	default:
		return false

	}
}

func (s *EcsService) runProxy(order *model.AliEcsOrder) {
	// 构建 代理连接信息
	tmpConnect := model.Connect{
		Id:      bson.NewObjectId().Hex(),
		Lname:   "阿里反向代理" + order.InstanceId,
		Saddr:   order.PublicIpAddress + ":22",
		User:    "root",
		Stype:   1,
		Passwd:  order.Password,
		Listen:  order.LocalListenAddr,
		Connect: "D",
		Active:  "Y",
	}
	if err := dal_connect.SaveData(&tmpConnect, true); err != nil {
		ulogs.Errorf("阿里云虚机自动添加动态代理失败 %v", err)
		return
	}
	db := dal.GetDB()
	tx := db.Model(model.AliEcsOrder{}).Where("id=?", order.Id).Update("connect_id", tmpConnect.Id)
	if err := tx.Error; err != nil {
		ulogs.Errorf("更新订单状态失败，更新订单隧道连接ID失败 %v", err)
	}
	if err := RunForwardInstance(tmpConnect.Id); err != nil {
		ulogs.Errorf("阿里云虚机自动添加动态代理失败 %v", err)
	}
}
