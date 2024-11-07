package iscsi

import (
	"bufio"
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

// 구조체 정의
type ISCSIInfo struct {
	TargetName          string
	CurrentPortal       string
	PersistentPortal    string
	Interface           ISCSIInterface
	Timeouts            ISCSITimeouts
	CHAP                ISCSIAuth
	NegotiatedParams    ISCSINegotiatedParams
	AttachedSCSIDevices []SCSIDevice
}

type ISCSIInterface struct {
	Name            string
	Transport       string
	Initiator       string
	IPAddress       string
	HWAddress       string
	Netdev          string
	SID             string
	ConnectionState string
	SessionState    string
}

type ISCSITimeouts struct {
	RecoveryTimeout    int
	TargetResetTimeout int
	LUNResetTimeout    int
	AbortTimeout       int
}

type ISCSIAuth struct {
	Username   string
	Password   string
	UsernameIn string
	PasswordIn string
}

type ISCSINegotiatedParams struct {
	HeaderDigest             string
	DataDigest               string
	MaxRecvDataSegmentLength int
	MaxXmitDataSegmentLength int
	FirstBurstLength         int
	MaxBurstLength           int
	ImmediateData            string
	InitialR2T               string
	MaxOutstandingR2T        int
}

type SCSIDevice struct {
	HostNumber  int
	State       string
	Protocol    string
	Channel     string
	ID          int
	Lun         int
	Device      string
	DeviceState string
}

func init() {
	log.SetLevel(log.DebugLevel)
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.RFC3339,
		NoColors:        true,
		CustomCallerFormatter: func(f *runtime.Frame) string {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf("[%s:%d %s()] ", path.Base(f.File), f.Line, funcName)
		},
	})
}

// 출력 파싱 함수
func parseISCSISession(output string) (*ISCSIInfo, error) {
	// 줄 단위로 입력 처리
	scanner := bufio.NewScanner(strings.NewReader(output))
	var iscsiInfo ISCSIInfo

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimLeft(line, " \t")

		// Target Name 추출
		if strings.HasPrefix(line, "Target:") {
			iscsiInfo.TargetName = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		// Current Portal 추출
		if strings.HasPrefix(line, "Current Portal:") {
			iscsiInfo.CurrentPortal = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		// Persistent Portal 추출
		if strings.HasPrefix(line, "Persistent Portal:") {
			iscsiInfo.PersistentPortal = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		// Interface 정보 추출
		if strings.HasPrefix(line, "Iface Name:") {
			iscsiInfo.Interface.Name = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Iface Transport:") {
			iscsiInfo.Interface.Transport = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Iface Initiatorname:") {
			iscsiInfo.Interface.Initiator = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Iface IPaddress:") {
			iscsiInfo.Interface.IPAddress = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Iface HWaddress:") {
			iscsiInfo.Interface.HWAddress = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Iface Netdev:") {
			iscsiInfo.Interface.Netdev = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "SID:") {
			iscsiInfo.Interface.SID = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "iSCSI Connection State:") {
			iscsiInfo.Interface.ConnectionState = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "iSCSI Session State:") {
			iscsiInfo.Interface.SessionState = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		// Timeouts 정보 추출
		if strings.HasPrefix(line, "Recovery Timeout:") {
			iscsiInfo.Timeouts.RecoveryTimeout = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "Target Reset Timeout:") {
			iscsiInfo.Timeouts.TargetResetTimeout = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "LUN Reset Timeout:") {
			iscsiInfo.Timeouts.LUNResetTimeout = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "Abort Timeout:") {
			iscsiInfo.Timeouts.AbortTimeout = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}

		// CHAP 정보 추출
		if strings.HasPrefix(line, "username:") {
			iscsiInfo.CHAP.Username = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "password:") {
			iscsiInfo.CHAP.Password = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "username_in:") {
			iscsiInfo.CHAP.UsernameIn = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "password_in:") {
			iscsiInfo.CHAP.PasswordIn = strings.TrimSpace(strings.Split(line, ":")[1])
		}

		// Negotiated iSCSI params 추출
		if strings.HasPrefix(line, "HeaderDigest:") {
			iscsiInfo.NegotiatedParams.HeaderDigest = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "DataDigest:") {
			iscsiInfo.NegotiatedParams.DataDigest = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "MaxRecvDataSegmentLength:") {
			iscsiInfo.NegotiatedParams.MaxRecvDataSegmentLength = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "MaxXmitDataSegmentLength:") {
			iscsiInfo.NegotiatedParams.MaxXmitDataSegmentLength = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "FirstBurstLength:") {
			iscsiInfo.NegotiatedParams.FirstBurstLength = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "MaxBurstLength:") {
			iscsiInfo.NegotiatedParams.MaxBurstLength = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}
		if strings.HasPrefix(line, "ImmediateData:") {
			iscsiInfo.NegotiatedParams.ImmediateData = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "InitialR2T:") {
			iscsiInfo.NegotiatedParams.InitialR2T = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "MaxOutstandingR2T:") {
			iscsiInfo.NegotiatedParams.MaxOutstandingR2T = parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
		}

		// Attached SCSI devices 추출
		if strings.HasPrefix(line, "Host Number:") {
			fields := strings.Fields(line)
			fmt.Print(fields)
			HostNumber := parseInt(strings.TrimSpace(strings.Split(line, ":")[1]))
			fmt.Print(HostNumber)
			for scanner.Scan() {
				var scsiDevice SCSIDevice
				line = scanner.Text()
				F := strings.Fields(line)
				scsiDevice.HostNumber = HostNumber
				scsiDevice.Protocol = strings.TrimSpace(F[0])
				scsiDevice.Channel = strings.TrimSpace(F[2])
				if scanner.Scan() {
					scsiDevice.ID = parseInt(F[4])
					scsiDevice.Lun = parseInt(F[6])
					line = scanner.Text()
					F = strings.Fields(line)
					scsiDevice.Device = F[3]
					scsiDevice.DeviceState = F[5]
				}
				iscsiInfo.AttachedSCSIDevices = append(iscsiInfo.AttachedSCSIDevices, scsiDevice)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &iscsiInfo, nil
}

// 문자열에서 정수를 추출하는 함수(융통성이 있음)
func parseInt(s string) int {
	var res int
	fmt.Sscanf(s, "%d", &res)
	return res
}

func parseAtoi(s string) int {
	str := strings.TrimSpace(s)
	num, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	} else {
		fmt.Println("Converted number:", num)
		return num
	}
}

func GetISCSISession() (*ISCSIInfo, error) {
	// iscsiadm 명령어 출력 예시
	cmd := exec.Command("iscsiadm", "-m", "session", "-P", "3")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error parsing output:", err)
		return nil, err
	}
	// 출력 파싱
	iscsiInfo, err := parseISCSISession(string(output))
	if err != nil {
		log.Error("Error parsing output:", err)
		return iscsiInfo, err
	}

	// 파싱된 정보 출력
	log.Debugf("%+v", iscsiInfo)
	return iscsiInfo, nil
}
