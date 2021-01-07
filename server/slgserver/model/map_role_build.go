package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/slgserver/proto"
	"slgserver/server/slgserver/static_conf"
	"slgserver/util"
	"time"
)

const (
	MapBuildFortress = 50
)

/*******db 操作begin********/
var dbRBMgr *rbDBMgr
func init() {
	dbRBMgr = &rbDBMgr{builds: make(chan *MapRoleBuild, 100)}
	go dbRBMgr.running()
}

type rbDBMgr struct {
	builds   chan *MapRoleBuild
}

func (this *rbDBMgr) running()  {
	for true {
		select {
		case b := <- this.builds:
			if b.Id >0 {
				_, err := db.MasterDB.Table(b).ID(b.Id).Cols(
					"rid", "name", "type", "level",
					"cur_durable", "max_durable", "occupy_time",
					"giveUp_time", "end_time").Update(b)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role build fail, because id <= 0")
			}
		}
	}
}

func (this *rbDBMgr) push(b *MapRoleBuild)  {
	this.builds <- b
}
/*******db 操作end********/

type MapRoleBuild struct {
	Id    		int    		`xorm:"id pk autoincr"`
	RId   		int    		`xorm:"rid"`
	Type  		int8   		`xorm:"type"`
	Level		int8   		`xorm:"level"`
	X          	int       	`xorm:"x"`
	Y          	int       	`xorm:"y"`
	Name       	string    	`xorm:"name"`
	Wood       	int       	`xorm:"-"`
	Iron       	int       	`xorm:"-"`
	Stone      	int       	`xorm:"-"`
	Grain      	int       	`xorm:"-"`
	CurDurable 	int       	`xorm:"cur_durable"`
	MaxDurable 	int       	`xorm:"max_durable"`
	Defender   	int       	`xorm:"defender"`
	OccupyTime 	time.Time 	`xorm:"occupy_time"`
	EndTime 	time.Time 	`xorm:"end_time"`	//建造完的时间
	GiveUpTime 	int64 		`xorm:"giveUp_time"`
}

func (this *MapRoleBuild) TableName() string {
	return "tb_map_role_build" + fmt.Sprintf("_%d", ServerId)
}

func (this* MapRoleBuild) LoadCfg() {
	if cfg, _ := static_conf.MapBuildConf.BuildConfig(this.Type, this.Level); cfg != nil {
		this.Wood = cfg.Wood
		this.Iron = cfg.Iron
		this.Stone = cfg.Stone
		this.Grain = cfg.Grain
	}
}


func (this* MapRoleBuild) Reset() {
	ok, t, level := MapResTypeLevel(this.X, this.Y)
	if ok {
		if cfg, _ := static_conf.MapBuildConf.BuildConfig(t, level); cfg != nil {
			this.Name = cfg.Name
			this.Level = cfg.Level
			this.Type = cfg.Type
			this.Wood = cfg.Wood
			this.Iron = cfg.Iron
			this.Stone = cfg.Stone
			this.Grain = cfg.Grain
		}
	}

	this.GiveUpTime = 0
	this.RId = 0
}

func (this* MapRoleBuild) IsInGiveUp() bool {
	return this.GiveUpTime != 0
}

func (this* MapRoleBuild) IsWarFree() bool  {
	curTime := time.Now().Unix()
	if curTime - this.OccupyTime.Unix() < static_conf.Basic.Build.WarFree{
		return true
	}else{
		return false
	}
}

func (this* MapRoleBuild) IsResBuild() bool  {
	return this.Grain > 0 || this.Stone > 0 || this.Iron > 0 || this.Wood > 0
}

//是否是要塞
func (this* MapRoleBuild) IsFortress() bool  {
	return this.Type == MapBuildFortress
}

func (this* MapRoleBuild) Build(cfg static_conf.BCLevelCfg) {
	this.Type = cfg.Type
	this.Level = 0
	this.Name = cfg.Name
	this.MaxDurable = cfg.Durable
	this.CurDurable = util.MinInt(this.MaxDurable, this.CurDurable)
	this.GiveUpTime = 0
	this.Defender = cfg.Defender
	this.Wood = 0
	this.Iron = 0
	this.Stone = 0
	this.Grain = 0
	this.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
}

/* 推送同步 begin */
func (this *MapRoleBuild) IsCellView() bool{
	return true
}

func (this *MapRoleBuild) IsCanView(rid, x, y int) bool{
	return true
}


func (this *MapRoleBuild) BelongToRId() []int{
	return []int{this.RId}
}

func (this *MapRoleBuild) PushMsgName() string{
	return "roleBuild.push"
}

func (this *MapRoleBuild) Position() (int, int){
	return this.X, this.Y
}

func (this *MapRoleBuild) TPosition() (int, int){
	return -1, -1
}

func (this *MapRoleBuild) ToProto() interface{}{

	p := proto.MapRoleBuild{}
	p.RNick = GetRoleNickName(this.RId)
	p.UnionId = GetUnionId(this.RId)
	p.UnionName = GetUnionName(p.UnionId)
	p.ParentId = GetParentId(this.RId)
	p.X = this.X
	p.Y = this.Y
	p.Type = this.Type
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable

	p.RId = this.RId
	p.Name = this.Name
	p.Defender = this.Defender
	p.OccupyTime = this.OccupyTime.UnixNano()/1e6
	p.GiveUpTime = this.GiveUpTime*1000
	p.EndTime = this.EndTime.UnixNano()/1e6

	if this.EndTime.IsZero() == false{
		if this.IsFortress(){
			if time.Now().Before(this.EndTime) == false{
				this.Level += 1
				this.EndTime = time.Time{}
			}
		}
	}
	p.Level = this.Level
	return p
}

func (this *MapRoleBuild) Push(){
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *MapRoleBuild) SyncExecute() {
	dbRBMgr.push(this)
	this.Push()
}