// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package services

import (
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestLoaderSReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.TemplatesCfg()["attrTemplateLoader"] = []*config.FCTemplate{
		{
			Type:  utils.MetaVariable,
			Path:  "*req.Accounts",
			Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
		},
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	conMngr := engine.NewConnManager(cfg, nil)
	srv := NewLoaderService(cfg, db, filterSChan,
		server, make(chan rpcclient.ClientConnector, 1),
		conMngr, anz, srvDep)
	srvMngr.AddServices(srv, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}

	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "loaders", "tutinternal"),
		Section: config.LoaderJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}

	if !db.IsRunning() {
		t.Fatal("Expected service to be running")
	}

	if !srv.IsRunning() {
		t.Fatal("Expected service to be running")
	}

	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	for _, v := range cfg.LoaderCfg() {
		v.Enabled = false
	}

	cfg.GetReloadChan(config.LoaderJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)

}