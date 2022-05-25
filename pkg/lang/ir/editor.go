// Copyright 2022 The envd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

import (
	"github.com/cockroachdb/errors"
	"github.com/moby/buildkit/client/llb"

	"github.com/tensorchord/envd/pkg/editor/vscode"
	"github.com/tensorchord/envd/pkg/flag"
	"github.com/tensorchord/envd/pkg/progress/compileui"
)

func (g Graph) compileVSCode() (*llb.State, error) {
	if len(g.VSCodePlugins) == 0 {
		return nil, nil
	}
	inputs := []llb.State{}
	for _, p := range g.VSCodePlugins {
		vscodeClient, err := vscode.NewClient(vscode.MarketplaceVendorOpenVSX)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create vscode client")
		}
		g.Writer.LogVSCodePlugin(p, compileui.ActionStart, false)
		if cached, err := vscodeClient.DownloadOrCache(p); err != nil {
			return nil, err
		} else {
			g.Writer.LogVSCodePlugin(p, compileui.ActionEnd, cached)
		}
		ext := llb.Scratch().File(llb.Copy(llb.Local(flag.FlagCacheDir),
			vscodeClient.PluginPath(p),
			"/home/envd/.vscode-server/extensions/"+p.String(),
			&llb.CopyInfo{
				CreateDestPath: true,
			}, llb.WithUIDGID(defaultUID, defaultGID)),
			llb.WithCustomNamef("install vscode plugin %s", p.String()))
		inputs = append(inputs, ext)
	}
	layer := llb.Merge(inputs, llb.WithCustomName("merging plugins for vscode"))
	return &layer, nil
}

func (g *Graph) compileJupyter() {
	if g.JupyterConfig != nil {
		g.PyPIPackages = append(g.PyPIPackages, "jupyter")
	}
}