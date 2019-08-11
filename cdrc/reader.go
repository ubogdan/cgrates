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

package cdrc

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type CDRcReader interface {
	ID() string                     // configuration identifier
	Init(args interface{}) error    // init will initialize the Reader, ie: open the file to read or http connection
	Read() (*utils.CGREvent, error) // Process a single record in the CDR file
	Processed() int64               // number of records processed
	Close() error                   // called when the reader should stop processing
}

// NewCDRCReader instantiates
func NewCDRcReader(cfg *config.CGRConfig, cfgIdx int) (cdrRdr CDRcReader, err error) {
	cfgPrfl := cfg.CDRcProfiles()[cfgIdx]
	switch cfgPrfl.CDRFormat {
	default:
		err = fmt.Errorf("unsupported CDR format: <%s>", cfgPrfl.CDRFormat)
	}
	return
}
