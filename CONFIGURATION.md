# flowpipeline Configuration and User Guide

Any flowpipeline is configured in a single yaml file which is either located in
the default `config.yml` or specified using the `-c` option when calling the
binary. The config file contains a single list of so-called segments, which
are processing flows in order. Flows represented by
[protobuf messages](https://github.com/bwNetFlow/protobuf/blob/master/flow-messages-enriched.proto)
within the pipeline.

Usually, the first segment is from the input group, followed by any number of
different segments. Often, flowpipelines end with a segment from the output,
print, or export groups. All segments, regardless from which group, accept and
forward their input from previous segment to their subsequent segment, i.e.
even input or output segments can be chained to one another or be placed in the
middle of a pipeline.

A list of full configuration examples with their own explanations can be found
[here](https://github.com/bwNetFlow/flowpipeline/tree/master/examples).

## Variable Expansion

Users can freely use environment variables in the `config` sections of any
segment. Additionally it is possible to reference command line arguments that
flowpipeline was invoked with that have not been parsed by the binary itself.
For instance:

```
segment: kafkaconsumer
config:
  user: myself
  pass: $PASSWORD
```

```
segment: flowfilter
config:
  filter: $0
```

```
segment: bpf
config:
  device: $0
```

## Available Segments

In addition to this section the detailed godoc can be used to get an overview
of available segments by going
[here](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline#section-directories)
and clicking `Expand all` in the bottom right.

### Alert Group
Segments in this group alert external endpoints via different channels. They're
usually placed behind some form of filtering and include data from flows making
it to this segment in their notification.

#### http
The `http` segment is currently a work in progress and is limited to sending
post requests with the full flow data to a single endpoint at the moment. The
roadmap includes various features such as different methods, builtin
conditional, limiting payload data, and multiple receivers.

```
- segment: http
  config:
    url: https://example.com/postable-endpoint
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/alert/http)
[examples using this segment](https://github.com/search?q=%22segment%3A+http%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Controlflow Group
Segments in this group have the ability to change the sequence of segments any
given flow traverses.

#### skip
The `skip` segment is used to conditionally skip over segments behind it. For
instance, in front of a export segment a condition such as `proto tcp` with a
skip value of `1` would result in any TCP flows not being exported by the
following segment. Setting invert to `true` is equivalent to negating the
condition.

```
- segment: skip
  config:
    condition: `proto tcp`
    # the lines below are optional and set to default
    skip: 1
    invert: false
```

[flowfilter syntax](https://github.com/bwNetFlow/flowfilter)
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/controlflow/skip)
[examples using this segment](https://github.com/search?q=%22segment%3A+skip%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Export Group
Segments in this group export flows to external databases. The distinction from
the output group lies in the fact that these exports are potentially lossy,
i.e. some fields might be lost. For instance, the `prometheus` segment as a
metric provider does not export any information about flow timing or duration,
among others.

#### prometheus
The `prometheus` segment provides a standard prometheus exporter, exporting its
own monitoring info at `:8080/metrics` and its flow data at `:8080/flowdata` by
default. The label set included with each metric is freely configurable with a
comma-separated list from any of the follwing valid fields: `router`,
`ipversion`, `application`, `protoname`, `direction`, `peer`, `remoteas`,
`remotecountry`, `src_port`, `dst_port`, `src_addr`, `dst_addr`.

Note that some of the above fields might not be present depending on the method
of flow export, the input segment used in this pipeline, or the modify segments
in front of this export segment.

```
- segment: prometheus
  config:
    labels: "router,ipversion,application,protoname,direction,peer,remoteas,remotecountry"
    # the lines below are optional and set to default
    endpoint: ":8080"
    metricspath: "/metrics"
    flowdatapath: "/flowdata"
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/export/prometheus)
[examples using this segment](https://github.com/search?q=%22segment%3A+prometheus%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Filter Group
Segments in this group all drop flows, i.e. remove them from the pipeline from
this segment on. Fields in individual flows are never modified, only used as
criteria.

#### flowfilter
The `flowfilter` segment uses
[flowfilter syntax](https://github.com/bwNetFlow/flowfilter) to drop flows
based on the evaluation value of the provided filter conditional against any
flow passing through this segment.

```
- segment: flowfilter
  config:
    filter: "proto tcp"
```

[flowfilter syntax](https://github.com/bwNetFlow/flowfilter)
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/filter)
[examples using this segment](https://github.com/search?q=%22segment%3A+filter%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Input Group
Segments in this group import or collect flows and provide them to all
following segments. As all other segments do, these still forward incoming
flows to the next segment, i.e. multiple input segments can be used in sequence
to add to a single pipeline.

#### bpf
The `bpf` segment sources packet header data from a local interface and uses
this data to run a Netflow-style cache before emitting flow data to the
pipeline.

This however has some caveats:
* using this segment requires root privileges to place BPF code in kernel
* the default kernel perf buffer size of 64kB should be sufficient on recent
  CPUs for up to 1Gbit/s of traffic, but requires tweaks in some scenarios
* the linux kernel version must be reasonably recent (probably 4.18+, certainly 5+)

Roadmap:
* allow hardware offloading to be configured
* implement sampling

```
- segment: bpf
  config:
    # required fields
    device: eth0
    # the lines below are optional and set to default
    activetimeout: 30m
    inactivetimeout: 15s
    buffersize: 65536
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/input/bpf)
[examples using this segment](https://github.com/search?q=%22segment%3A+bpf%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### goflow
The `goflow` segment provides a convenient interface for
[goflow2](https://github.com/netsampler/goflow2) right from flowpipeline
config. It behaves closely like a regular goflow2 instance but provides flows
directly in our extended format (which has been based on goflow's from the
beginning of this project).

This flow collector needs to receive input from any IPFIX/Netflow/sFlow
exporters, for instance your network devices.

```
- segment: goflow
  # the lines below are optional and set to default
  config:
    listen: "sflow://:6343,netflow://:2055"
    workers: 1
```

[goflow2 fields](https://github.com/netsampler/goflow2/blob/main/docs/protocols.md)
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/input/goflow)
[examples using this segment](https://github.com/search?q=%22segment%3A+goflow%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### kafkaconsumer
The `kafkaconsumer` segment consumes flows from a Kafka topic. This topic can
be created using the `kafkaproducer` module or using an external instance of
[goflow2](https://github.com/netsampler/goflow2).

This segment can be used in conjunction with the `kafkaproducer` segment to
enrich, reduce, or filter flows in transit between Kafka topics, or even sort
them into different Kafka topics. See the examples this particular usage.

```
- segment: kafkaconsumer
  config:
    # required fields
    server: some.kafka.server.example.com:9092
    topic: flow-topic-name
    group: consumer-group-name
    tls: true
    auth: true
    # required if auth is true
    user: myusername
    pass: mypassword
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/input/kafkaconsumer)
[examples using this segment](https://github.com/search?q=%22segment%3A+kafkaconsumer%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

##### BelWü-connected Entities
If you are connected behind [BelWü](https://www.belwue.de) contact one of our
developers or the [bwNet project](bwnet.belwue.de) for access to our central
Kafka cluster. We'll gladly create a topic for your university if it doesn't
exist already and provide you with the proper configuration to use this
segment.

If the person responsible at university agrees, we will pre-fill this topic
with flows sent or received by your university as seen on our upstream border
interfaces. This can be limited according to the data protection requirements
set forth by the universities.

#### stdin
The `stdin` segment reads JSON encoded flows from stdin and introduces this
into the pipeline. This is intended to be used in conjunction with the `stdout`
segment, which allows flowpipelines to be piped into each other.

```
- segment: stdin
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/input/stdin)
[examples using this segment](https://github.com/search?q=%22segment%3A+stdin%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Modify Group
Segments in this group modify flows in some way. Generally, these segments do
not drop flows unless specifically instructed and only change fields within
them. This group contains both, enriching and reducing segments.

#### addcid
The `addcid` segment can add a customer ID to flows according to the IP prefix
the flow is matched to. These prefixes are sourced from a simple csv file
consisting of lines in the format `ip prefix,integer`. For example:

```csv
192.168.88.0/24,1
2001:db8:1::/48,1
```

Which IP address is matched against this data base is determined by the
RemoteAddress field of the flow. If this is unset, the flow is forwarded
untouched. To set this field, see the `remoteaddress` segment. If matchboth is
set to true, this segment will not try to establish the remote address and
instead check both, source and destination address, in this order on a
first-match basis. The assumption here is that there are no flows of customers
sending traffic to one another.

If dropunmatched is set to true no untouched flows will pass this segment,
regardless of the reason for the flow being unmatched (absence of RemoteAddress
field, actually no matching entry in data base).

Roadmap:
* figure out how to deal with customers talking to one another

```
- segment: addcid
  config:
    filename: filename.csv
    # the lines below are optional and set to default
    dropunmatched: false
    matchboth: false
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/addcid)
[examples using this segment](https://github.com/search?q=%22segment%3A+addcid%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### anonymize
The `anonymize` segment anonymizes IP addresses occuring in flows using the
Crypto-PAn algorithm. By default all possible IP address fields are targeted,
this can be configured using the fields parameter.

```
- segment: anonymize
  config:
    encryptionkey: "abcdef"
    # the lines below are optional and set to default
    fields: "SrcAddr,DstAddr,SamplerAddress"
```

[any additional links](https://bwnet.belwue.de)
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/anonymize)
[examples using this segment](https://github.com/search?q=%22segment%3A+anonymize%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### dropfields
The segment `dropfields` deletes fields from flows as they pass through this
segment. To this end, this segment requires a policy parameter to be set to
"keep" or "drop". It will then either keep or drop all fields specified in the
fields parameter. For a list of fields, check our
[protobuf definition](https://github.com/bwNetFlow/protobuf/blob/master/flow-messages-enriched.proto).

```
- segment: dropfields
  config:
    # required, options are drop or keep
    policy: drop
    # the lines below are optional and set to default
    fields: ""
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/dropfields)
[examples using this segment](https://github.com/search?q=%22segment%3A+dropfields%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### geolocation
The `geolocation` segment annotates flows with their RemoteCountry field.
Requires the filename parameter to be set to the location of a MaxMind
geolocation file, as shown in our example.

For this to work, it requires RemoteAddress to be set in the flow if matchboth
is set to its default `false`. If matchboth is true, the behaviour is
different from the `addcid` segment, as the result will be written for both
SrcAddr and DstAddr into SrcCountry and DstCountry. The dropunmatched parameter
however behaves in the same way: flows without any remote country data set will
be dropped.

```
- segment: geolocation
  config:
    # required
    filename: file.mmdb
    # the lines below are optional and set to default
    dropunmatched: false
    matchboth: false
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/geolocation)
[examples using this segment](https://github.com/search?q=%22segment%3A+geolocation%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### normalize
The `normalize` segment multiplies the Bytes and the Packets field by the flows
SamplingRate field. Additionally, it sets the Normalized field for this flow to 1.

The fallback parameter is for flows known to be sampled which do not include
the sampling rate for some reason.

Roadmap:
* replace Normalized with an OriginalSamplingRate field and set SamplingRate to 1 instead

```
- segment: normalize
  # the lines below are optional and set to default
  config:
    fallback: 0
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/normalize)
[examples using this segment](https://github.com/search?q=%22segment%3A+normalize%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### protomap
The `protomap` segment sets the ProtoName string field according to the Proto
integer field. Note that this should only be done before final usage, as
lugging additional string content around can be costly regarding performance
and storage size.

```
- segment: protomap
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/protomap)
[examples using this segment](https://github.com/search?q=%22segment%3A+protomap%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### remoteaddress
The `remoteaddress` segment determines any given flows remote address and sets
the RemoteAddress field up to indicate either SrcAddr or DstAddr as the remote
address. This is done by different policies, see
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/remoteaddress)
for a detailed explanation.

The short version is:
* `cidr` assumes the remote address is the one that has a match in the CSV file
  used by the `addcid` segment. Done for source and destination addres.
* `border` assumes all flows originate at the outside network border, i.e. on
  peering, exchange, or transit interfaces. Thus, any incoming flows originate
  at a remote address (source address), and any outgoing flows originate at a
  local address (destination address is remote).
* `user` assumes the opposite, namely that flows are collected on user- or
  customer-facing interfaces. The assignment is thus reversed.
* `clear` clears all remote address info and is thus equivalent to using the
  `dropfields` segment. This can be used when processing flows from mixed
  sources before reestablishing remote address using `cidr`.

Any optional parameters relate to the `cidr` policy only and behave as in the
`addcid` segment.

```
- segment: remoteaddress
  config:
    # required, one of cidr, border, user, or clear
    policy: cidr
    # required if policy is cidr
    filename: same_csv_file_as_for_addcid_segment.csv
    # the lines below are optional and set to default, relevant to policy cidr only
    dropunmatched: false
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/remoteaddress)
[examples using this segment](https://github.com/search?q=%22segment%3A+remoteaddress%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### snmp
The `snmp` segment annotates flows with information learned directly from
routers using SNMP. This is a potentially perfomance impacting segment and has
to be configured carefully.

In principle, this segment tries to fetch a SNMP OID datapoint from the address
in SamplerAddress, which corresponds to a router on normal flow-exporter
generated flows. The fields used to query this router are SrcIf and DstIf, i.e.
the interface IDs which are part of the flow. If successfull, a flow will have
the fields `{Src,Dst}IfName`, `{Src,Dst}IfDesc`, and `{Src,Dst}IfSpeed`
populated. In order to not to overload the router and to introduce delays, this
segment will:

* not wait for a SNMP query to return, instead it will leave the flow as it was
  before sending it to the next segment (i.e. the first one on a given
  interface will always remain untouched)
* add any interface's data to a cache, which will be used to enrich the
  next flow using that same interface
* clear the cache value after 1 hour has elapsed, resulting in another flow
  without these annotations at that time

These rules are applied for source and destination interfaces separately.

The paramters to this segment specify the SNMPv2 community as well as the
connection limit employed by this segment. The latter is again to not overload
the routers SNMPd. Lastly, the regex parameter can be used to limit the
`IfDesc` annotations to a certain part of the actual interface description.
For instance, descriptions follow the format `customerid - blablalba`, the
regex `(.*) -.*` would grab just that customer ID to put into the `IfDesc`
fields. Also see the full examples linked below.

Roadmap:
* cache timeout should be configurable

```
- segment: snmp
  # the lines below are optional and set to default
  config:
    community: public
    regex: ".*"
    connlimit: 16

```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/modify/snmp)
[examples using this segment](https://github.com/search?q=%22segment%3A+snmp%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Output Group

#### kafkaproducer
The `kafkaproducer` segment produces flows to a Kafka topic. All settings are
equivalent to the `kafkaconsumer` segment. Additionally, there is the
`topicsuffix` parameter, which allows pipelines to write to multiple topics at
once. A typical use case is this:
* set `topic: customer-`
* set `topicsuffix: Cid`
* this will result in a number of topics which will a) be created or b) need to
  exist, depending on your Kafka cluster settings.
* the topics will take the form `customer-123` for all values of Cid
* it is advisable to use the `flowfilter` segment to limit the number of
  topics

This could also be used to populate topics by Proto, or by Etype, or by any
number of other things.

```
- segment: kafkaproducer
  config:
    # required fields
    server: some.kafka.server.example.com:9092
    topic: flow-topic-name
    group: consumer-group-name
    tls: true
    auth: true
    # required if auth is true
    user: myusername
    pass: mypassword
    # the lines below are optional and set to default
    topicsuffix: ""
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/output/kafkaproducer)
[examples using this segment](https://github.com/search?q=%22segment%3A+kafkaproducer%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### sqlite
TODO: major rewrite incoming with MR #18 output/csv

```
- segment: sqlite
  config:

```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/output/sqlite)
[examples using this segment](https://github.com/search?q=%22segment%3A+sqlite%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### stdout
The `stdout` segment exports all flows passing it as JSON on stdout. This is
intended to be able to pipe flows between instances of flowpipeline, but it is
also very useful when debugging flowpipelines or to create a quick plaintext
dump.

```
- segment: stdout
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/output/stdout)
[examples using this segment](https://github.com/search?q=%22segment%3A+stdout%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Print Group
Segments in this group serve to print flows immediately to the user. This is intended for ad-hoc applications and instant feedback use cases.

#### count
The `count` segment counts flows passing it. This is mainly for debugging
flowpipelines. For instance, placing two of these segments around a
`flowfilter` segment allows users to use the `prefix` parameter with values
`"pre-filter: "`  and `"post-filter: "` to obtain a count of flows making it
through the filter without resorting to some command employing `| wc -l`.

The result is printed upon termination of the flowpipeline.

```
- segment: count
  # the lines below are optional and set to default
  config:
    prefix: ""
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/print/count)
[examples using this segment](https://github.com/search?q=%22segment%3A+count%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### printdots
The `printdots` segment keeps counting flows internally and emits a dot (`.`)
every `flowsperdot` flows. Its parameter needs to be chosen with the expected
flows per second in mind to be useful. Used to get visual feedback when
necessary.

```
- segment: printdots
  # the lines below are optional and set to default
  config:
    flowsperdot: 5000
```

[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/print/printdots)
[examples using this segment](https://github.com/search?q=%22segment%3A+printdots%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

#### printflowdump
The `printflowdump` prints a tcpdump-style representation of flows with some
addition deemed useful at [BelWü](https://www.belwue.de) Ops. It looks up
protocol names from flows if the segment `protoname` is part of the pipeline so
far or determines them on its own (using the same method as the `protoname`
segment), unless configured otherwise.

It currently looks like `timereceived: SrcAddr -> DstAddr [ingress iface desc from snmp segment →
@SamplerAddress → egress iface desc], ProtoName, Duration, avg bps, avg pps`. In
action:

```
14:47:41: x.x.x.x:443 -> 193.197.x.x:9854 [Telia → @193.196.190.193 → stu-nwz-a99], TCP, 52s, 52.015384 Mbps, 4.334 kpps
14:47:49: 193.197.x.x:54643 -> x.x.x.x:7221 [Uni-Ulm → @193.196.190.2 → Telia], UDP, 60s, 2.0288 Mbps, 190 pps
14:47:49: 2003:x:x:x:x:x:x:x:51052 -> 2001:7c0:x:x::x:443 [DTAG → @193.196.190.193 → stu-nwz-a99], UDP, 60s, 29.215333 Mbps, 2.463 kpps
```

This segment is commonly used as a pipeline with some input provider and the
`flowfilter` segment preceeding it, to create a tcpdump-style utility. This is
even more versatile when using `$0` as a placeholder in `flowfilter` to use the
flowpipeline invocation's first argument as a filter.

```
- segment: printflowdump
  # the lines below are optional and set to default
  config:
    useprotoname: true

```
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/print/printflowdump)
[examples using this segment](https://github.com/search?q=%22segment%3A+printflowdump%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)

### Ungrouped

This is for internally used segments only.

#### noop
The `noop` segment serves as a heavily annotated template for new segments. So
does this piece of documentation. Aside from summarizing what a segment does,
it should include a description of all the parameters it accepts as well as any
caveats users should be aware of.

Roadmap:
* things to expect go here
* and here

```
- segment: noop
  # the lines below are optional and set to default
  config:
    jk: this segment actually has no config at all, its just for this template
```

[any additional links](https://bwnet.belwue.de)
[godoc](https://pkg.go.dev/github.com/bwNetFlow/flowpipeline/segments/noop)
[examples using this segment](https://github.com/search?q=%22segment%3A+noop%22+extension%3Ayml+repo%3AbwNetFlow%2Fflowpipeline%2Fexamples&type=Code)
