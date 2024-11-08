package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestSuite 정의
type TestSuite struct {
	suite.Suite
	tr Trace
}

// TestCalculatorTestSuite: 스위트를 실행하는 메인 테스트 함수
func TestCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// SetupTest: 각 테스트 전에 실행될 설정 메서드
func (s *TestSuite) SetupTest() {
	var err error
	targetCommand := "fio"
	s.tr, err = NewTrace(&targetCommand)
	assert.NotNil(s.T(), err)
}

// TestAdd: Add 메서드 테스트
func (s *TestSuite) Test_UpdateNodeGraph() {
	err := s.tr.CreateNodeGraphData()
	assert.NotNil(s.T(), err)
	//assert.Equal(s.T(), 5, result, "2 + 3은 5여야 합니다.")
	err = s.tr.CreatePrometheusMetric()
	assert.NotNil(s.T(), err)

}
