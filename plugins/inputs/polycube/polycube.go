package polycube

import (
    "fmt"
    "strings"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
    dmclient "github.com/goldenrye/go-polycube/ddosmitigator"
)

type PolycubeStats struct {
    dmStatus        map[string]bool     // whether ddosmitigator active or not
    lastDMStats     map[string]uint64

    Node                string      `toml:"node"`
    Ddosmitigatorlist   string      `toml:"ddosmitigatorlist"`
    Basepath            string      `toml:"basepath"`
    Enable              bool        `toml:"enable"`
}

var sampleConfig = `
    ## Whether enable the stats collecting
    enable = true
    ## Which node the stats comes from
    node = "polycube-dev"
    ## Which DdosMitigator to collect stats
    ddosmitigatorlist = "dm1,dm2"
    ## Base path polycube Restful API server
    basepath = "http://localhost:9000/polycube/v1"
`

func (_ *PolycubeStats) SampleConfig() string {
    return sampleConfig
}

func (pc *PolycubeStats) Description() string {
    return "Read polycube metrics"
}

func (pc *PolycubeStats) Gather(acc telegraf.Accumulator) error {
    if !pc.Enable {
        return nil
    }
    cfg := dmclient.Configuration{
        BasePath: pc.Basepath,
    }
    cli := dmclient.NewAPIClient(&cfg)

    dmlist := strings.Split(pc.Ddosmitigatorlist, ",")
    for _, dm := range dmlist {
        sbl_list, _, err := cli.DdosmitigatorApi.ReadDdosmitigatorBlacklistSrcListByID(nil, dm)
        if err != nil {
            continue
        }
        for _, sbl := range sbl_list {
            fmt.Printf("%s: %d\n", sbl.Ip, sbl.DropPkts)
        }
    }

    return nil
}

func init() {
    inputs.Add("polycube", func() telegraf.Input {
        return &PolycubeStats{
        }
    })
}
