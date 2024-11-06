## 기본 구성 
### Node exporter 
#### 설치와 기동 
```sh

wget https://github.com/prometheus/node_exporter/releases/download/v1.3.1/node_exporter-1.8.2.linux-amd64.tar.gz
wget https://github.com/prometheus/node_exporter/releases/download/v1.3.1/node_exporter-1.3.1.linux-arm64.tar.gz
 
==> 입맛에 맛는 버젼으로 선택
node_exporter-1.8.2.linux-amd64.tar.gz
tar xvfz node_exporter-1.8.2.linux-amd64.tar.gz
nohup ./node_exporter&
ss -natp
```

###  Prometheus 
* 설치
```sh
wget https://github.com/prometheus/prometheus/releases/download/v2.54.0-rc.1/prometheus-2.54.0-rc.1.linux-amd64.tar.gz --no-check-certificate

==> 입맛에 맛는 버젼으로 선택
tar xvfz prometheus-2.52.0.linux-amd64.tar.gz
==> scrape_config
 
nohup ./prometheus --config.file=prometheus.yml&
 
==> expert 상태 확인
http://192.168.137.30:9090/targets
```
* prometheus.yml 설정
```sh
$ cat prometheus.yml 
# my global config
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093

# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"


# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: "node-g101"
    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
    static_configs:
      - targets: ["192.168.137.101:9100"] 
  - job_name: "node-g102"
    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
    static_configs:
      - targets: ["192.168.137.102:9100"] 
  - job_name: "node-g103"
    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
    static_configs:
      - targets: ["192.168.137.103:9100"]      
  - job_name: "node-grph-g103"
    static_configs:
      - targets: ["192.168.137.103:2224"]
```

### Grafana
* 설치
```sh
wget https://dl.grafana.com/enterprise/release/grafana-enterprise-11.1.0.linux-amd64.tar.gz
wget https://dl.grafana.com/enterprise/release/grafana-enterprise-9.0.5.linux-arm64.tar.gz
 
grafana-enterprise-11.1.0.linux-amd64.tar.gz
 
$ cat defaults.ini | grep  http_port
http_port = 3000
 
nohup ./grafana-server&

http://192.168.137.30:3000
```
node export full 설정
https://grafana.com/grafana/dashboards/1860-node-exporter-full/

적용 방법은 페이지에서 Copy id to clipboard를 통해 id를 복사하고, Grafana에서 Dashboards -> +import에 진입하여



## System 데몬 설정
### prometheus 
```sh
$ wget https://github.com/prometheus/prometheus/releases/download/v2.33.4/prometheus-2.33.4.linux-amd64.tar.gz
$ wget https://github.com/prometheus/node_exporter/releases/download/v1.3.1/node_exporter-1.3.1.linux-amd64.tar.gz
$ gunzip *.gz
$ tar xvf node_exporter-1.3.1.linux-amd64.tar
$ tar xvf prometheus-2.33.4.linux-amd64.tar
 
$ sudo mkdir -p /usr/local/prometheus
$ sudo mv console_libraries/ consoles/ prometheus /usr/local/prometheus/
$ sudo mv prometheus.yml /usr/local/prometheus/
 
$ sudo groupadd --system prometheus
$ sudo useradd --system -s /usr/sbin/nologin -g prometheus prometheus
 
$ sudo  chown -R  prometheus:prometheus /usr/local/prometheus
# mkdir  -p /var/lib/prometheus/
# chown  prometheus:prometheus /var/lib/prometheus/
 
cat /etc/systemd/system/prometheus.service
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target
 
[Service]
User=prometheus
Restart=on-failure
ExecStart=/usr/local/prometheus/prometheus \
    --config.file=/usr/local/prometheus/prometheus.yml \
    --storage.tsdb.path=/var/lib/prometheus/ \
    --web.console.templates=/usr/prometheus/console \
    --web.console.libraries=/usr/prometheus/console_libraries \
    --web.listen-address=0.0.0.0:9090 \
    --web.external-url=
 
[Install]
WantedBy=multi-user.target
 
$ systemctl daemon-reload
$ systemctl start prometheus
$ journalctl -xe
 
* 확인:  http://localhost:9090
```
* prometheus.yml

```yaml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: "prometheus"
 
    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
 
    static_configs:
      - targets: ["localhost:9090"]
 
  - job_name: "node_exporter"
    static_configs:
      - targets: ["localhost:9100"]
```

### Node Exporter 

```sh

# mkdir -p  /usr/local/prometheus/exporters
# chown -R prometheus:prometheus /usr/local/prometheus/exporters/
$ sudo mv node_exporter  /usr/local/prometheus/exporters/
```
 
```sh
# cat  node_exporter.service
[Unit]
Description=Prometheus - node_exporter
Wants=network-online.target
After=network-online.target
 
[Service]
User=prometheus
Restart=on-failure
ExecStart=/usr/local/prometheus/exporters/node_exporter
 
[Install]
WantedBy=multi-user.target
 
# systemctl daemon-reload
# systemctl start node_exporter.service
# systemctl status  node_exporter.service
$ journalctl -xe
 
 
* 확인: http://localhost:9100/metrics
```

### grafana
```sh
$ sudo apt-get install -y apt-transport-https
$ sudo apt-get install -y software-properties-common wget
$ wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
$ echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list
$ sudo apt-get update
$ sudo apt-get install grafana
 
$ sudo systemctl daemon-reload
$ sudo systemctl enable grafana-server
$ sudo systemctl start grafana-server
 
# cat grafana-server.service
[Unit]
Description=Grafana instance
Documentation=http://docs.grafana.org
Wants=network-online.target
After=network-online.target
After=postgresql.service mariadb.service mysql.service
 
[Service]
EnvironmentFile=/etc/default/grafana-server
User=grafana
Group=grafana
Type=simple
Restart=on-failure
WorkingDirectory=/usr/share/grafana
RuntimeDirectory=grafana
RuntimeDirectoryMode=0750
ExecStart=/usr/sbin/grafana-server                                                  \
                            --config=${CONF_FILE}                                   \
                            --pidfile=${PID_FILE_DIR}/grafana-server.pid            \
                            --packaging=deb                                         \
                            cfg:default.paths.logs=${LOG_DIR}                       \
                            cfg:default.paths.data=${DATA_DIR}                      \
                            cfg:default.paths.plugins=${PLUGINS_DIR}                \
                            cfg:default.paths.provisioning=${PROVISIONING_CFG_DIR} 
 
 
LimitNOFILE=10000
TimeoutStopSec=20
CapabilityBoundingSet=
DeviceAllow=
LockPersonality=true
MemoryDenyWriteExecute=false
NoNewPrivileges=true
PrivateDevices=true
PrivateTmp=true
ProtectClock=true
ProtectControlGroups=true
ProtectHome=true
ProtectHostname=true
ProtectKernelLogs=true
ProtectKernelModules=true
ProtectKernelTunables=true
ProtectProc=invisible
ProtectSystem=full
RemoveIPC=true
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
RestrictNamespaces=true
RestrictRealtime=true
RestrictSUIDSGID=true
SystemCallArchitectures=native
UMask=0027
 
[Install]
WantedBy=multi-user.target
 
* 확인 : localhost:3000 admin/admin
```