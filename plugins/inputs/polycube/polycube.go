package polycube

import (
    "fmt"
    "strings"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
    dmclient "github.com/goldenrye/go-polycube/ddosmitigator"
)

type PolycubeStats struct {
    lastDMDropPkts     map[string]uint64   // key: node|dm|<src|dst>|ip

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
            key := pc.Node + "|" + dm + "|src|" + sbl.Ip
            _, ok := pc.lastDMDropPkts[key]; if !ok {
                pc.lastDMDropPkts[key] = (uint64)(sbl.DropPkts)
                continue
            }
            dropPkts := (uint64)(sbl.DropPkts) - pc.lastDMDropPkts[key]
            if dropPkts == 0 {
                continue
            }
            fmt.Printf("%s: %d\n", key, dropPkts)
            pc.lastDMDropPkts[key] = (uint64)(sbl.DropPkts)

            fields := map[string]interface{}{
                "drop_pkts":    dropPkts,
            }
            tags := map[string]string{
                "node": pc.Node,
                "dm":   dm,
                "dir":  "src",
                "ip":   sbl.Ip,
            }
            acc.AddGauge("ddos", fields, tags)
        }
    }

    return nil
}

func init() {
    inputs.Add("polycube", func() telegraf.Input {
        return &PolycubeStats{
            lastDMDropPkts : make(map[string]uint64),
        }
    })
}
