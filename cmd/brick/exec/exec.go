package exec

import (
	"sort"
	"strings"

	"github.com/aergoio/aergo-lib/log"

	"github.com/aergoio/aergo/cmd/brick/context"
)

var logger = log.NewLogger("brick")

type Executor interface {
	Command() string
	Syntax() string
	Usage() string
	Describe() string
	Validate(args string) error
	Run(args string) (string, error)
}

var execImpls = make(map[string]Executor)

func registerExec(executor Executor) {
	execImpls[executor.Command()] = executor
	Index(context.CommandSymbol, executor.Command())
}

func GetExecutor(cmd string) Executor {
	return execImpls[cmd]
}

func AllExecutors() []Executor {

	var result []Executor

	for _, executor := range execImpls {
		result = append(result, executor)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].Command(), result[j].Command()) < 0
	})

	return result
}

func Broker(cmdStr string) {

	cmd, args := context.ParseFirstWord(cmdStr)
	if len(cmd) == 0 || context.Comment == cmd {
		return
	}

	Execute(cmd, args)
}

func Execute(cmd, args string) {
	executor := GetExecutor(cmd)

	if executor == nil {
		logger.Warn().Str("cmd", cmd).Msg("command not found")
		return
	}

	if err := executor.Validate(args); err != nil {
		logger.Error().Err(err).Str("cmd", cmd).Msg("validation fail")
		return
	}

	result, err := executor.Run(args)
	if err != nil {
		logger.Error().Err(err).Str("cmd", cmd).Msg("execution fail")
		return
	}

	//logger.Info().Str("result", result).Str("cmd", cmd).Msg("execution success")
	logger.Info().Str("cmd", cmd).Msg(result)
}
