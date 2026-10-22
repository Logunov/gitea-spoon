/*
Copyright 2024 Dan Molik.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package giteadb

import (
	"context"
	"fmt"
	"sync"

	"gitea.dev/models/db"
	"gitea.dev/modules/setting"
)

var (
	initMu           sync.Mutex
	initialized      bool
	initializeEngine = initEngine
)

// Init initializes Gitea's process-wide database engine at most once after a
// successful initialization. Failed attempts are not cached, so callers may
// retry when the database or configuration becomes available.
func Init(ctx context.Context) error {
	initMu.Lock()
	defer initMu.Unlock()

	if initialized {
		return nil
	}

	if err := initializeEngine(ctx); err != nil {
		return err
	}

	initialized = true
	return nil
}

func initEngine(ctx context.Context) error {
	setting.InitCfgProvider("/data/gitea/conf/app.ini")
	setting.LoadDBSetting()

	if setting.Database.Type == "" {
		return fmt.Errorf(`database settings are missing from the configuration file: %q
Ensure you are running in the correct environment or set the correct configuration file with -c.
If this is the intended configuration file complete the [database] section`, setting.CustomConf)
	}

	if err := db.InitEngine(ctx); err != nil {
		return fmt.Errorf("unable to initialize the database using the configuration in %q: %w", setting.CustomConf, err)
	}

	return nil
}
