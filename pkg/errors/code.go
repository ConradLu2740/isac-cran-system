package errors

type Code int

const (
	CodeSuccess Code = 0

	CodeBadRequest         Code = 400
	CodeUnauthorized       Code = 401
	CodeForbidden          Code = 403
	CodeNotFound           Code = 404
	CodeInternalError      Code = 500
	CodeServiceUnavailable Code = 503

	CodeInvalidParam     Code = 10001
	CodeInvalidIRSConfig Code = 10002
	CodeInvalidChannelID Code = 10003
	CodeInvalidAlgorithm Code = 10004

	CodeIRSDeviceError  Code = 20001
	CodeIRSConfigFailed Code = 20002
	CodeIRSStatusError  Code = 20003

	CodeUSRPDeviceError  Code = 21001
	CodeUSRPReceiveError Code = 21002

	CodeSensorConnectError Code = 22001
	CodeSensorDataError    Code = 22002

	CodeAlgorithmRunError   Code = 30001
	CodeAlgorithmTimeout    Code = 30002
	CodeAlgorithmNoConverge Code = 30003

	CodeDBConnectError Code = 40001
	CodeDBQueryError   Code = 40002
	CodeDBInsertError  Code = 40003
	CodeDBUpdateError  Code = 40004

	CodeRedisConnectError Code = 41001
	CodeRedisOpError      Code = 41002

	CodeInfluxConnectError Code = 42001
	CodeInfluxQueryError   Code = 42002
	CodeInfluxWriteError   Code = 42003

	CodeMQTTConnectError Code = 43001
	CodeMQTTPublishError Code = 43002

	CodeMATLABExportError Code = 50001
	CodeMATLABImportError Code = 50002

	CodeExperimentNotFound Code = 60001
	CodeExperimentRunning  Code = 60002
)

var codeMessages = map[Code]string{
	CodeSuccess:            "success",
	CodeBadRequest:         "bad request",
	CodeUnauthorized:       "unauthorized",
	CodeForbidden:          "forbidden",
	CodeNotFound:           "not found",
	CodeInternalError:      "internal server error",
	CodeServiceUnavailable: "service unavailable",

	CodeInvalidParam:     "invalid parameter",
	CodeInvalidIRSConfig: "invalid IRS configuration",
	CodeInvalidChannelID: "invalid channel ID",
	CodeInvalidAlgorithm: "invalid algorithm type",

	CodeIRSDeviceError:  "IRS device error",
	CodeIRSConfigFailed: "IRS configuration failed",
	CodeIRSStatusError:  "IRS status error",

	CodeUSRPDeviceError:  "USRP device error",
	CodeUSRPReceiveError: "USRP receive error",

	CodeSensorConnectError: "sensor connection error",
	CodeSensorDataError:    "sensor data error",

	CodeAlgorithmRunError:   "algorithm run error",
	CodeAlgorithmTimeout:    "algorithm timeout",
	CodeAlgorithmNoConverge: "algorithm did not converge",

	CodeDBConnectError: "database connection error",
	CodeDBQueryError:   "database query error",
	CodeDBInsertError:  "database insert error",
	CodeDBUpdateError:  "database update error",

	CodeRedisConnectError: "redis connection error",
	CodeRedisOpError:      "redis operation error",

	CodeInfluxConnectError: "influxdb connection error",
	CodeInfluxQueryError:   "influxdb query error",
	CodeInfluxWriteError:   "influxdb write error",

	CodeMQTTConnectError: "mqtt connection error",
	CodeMQTTPublishError: "mqtt publish error",

	CodeMATLABExportError: "matlab export error",
	CodeMATLABImportError: "matlab import error",

	CodeExperimentNotFound: "experiment not found",
	CodeExperimentRunning:  "experiment is running",
}

func (c Code) Message() string {
	if msg, ok := codeMessages[c]; ok {
		return msg
	}
	return "unknown error"
}

func (c Code) Int() int {
	return int(c)
}
