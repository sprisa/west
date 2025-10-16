module github.com/sprisa/west

go 1.25.2

tool (
	github.com/99designs/gqlgen
	github.com/Khan/genqlient
)

require (
	entgo.io/contrib v0.7.0
	entgo.io/ent v0.14.5
	github.com/99designs/gqlgen v0.17.81
	github.com/Khan/genqlient v0.8.1
	github.com/anandvarma/namegen v1.1.1
	github.com/cqroot/prompt v0.9.4
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/miekg/dns v1.1.61
	github.com/rs/zerolog v1.34.0
	github.com/samber/lo v1.52.0
	github.com/sirupsen/logrus v1.9.3
	github.com/slackhq/nebula v1.9.7
	github.com/urfave/cli/v3 v3.4.1
	github.com/vektah/gqlparser/v2 v2.5.30
	golang.org/x/crypto v0.42.0
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b
	golang.org/x/sync v0.17.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.39.1
)

require (
	ariga.io/atlas v0.32.1-0.20250325101103-175b25e1c1b9 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alexflint/go-arg v1.5.1 // indirect
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/charmbracelet/bubbles v0.16.1 // indirect
	github.com/charmbracelet/bubbletea v0.24.2 // indirect
	github.com/charmbracelet/lipgloss v0.11.0 // indirect
	github.com/charmbracelet/x/ansi v0.1.1 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/cqroot/multichoose v0.1.1 // indirect
	github.com/cyberdelia/go-metrics-graphite v0.0.0-20161219230853-39f87cc3b432 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/flynn/noise v1.1.0 // indirect
	github.com/gaissmai/bart v0.11.1 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl/v2 v2.18.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/nbrownus/go-metrics-prometheus v0.0.0-20210712211119-974a6260965f // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/vishvananda/netlink v1.2.1-beta.2 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20230325221338-052af4a8072b // indirect
	golang.zx2c4.com/wireguard/windows v0.5.3 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
