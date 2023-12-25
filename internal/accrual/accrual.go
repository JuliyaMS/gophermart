package accrual

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"
)

type SystemAccrual struct {
	sem  *Semaphore
	conn *DBAccrual
	log  *zap.SugaredLogger
}

func NewSystemAccrual(logger *zap.SugaredLogger, maxSem int) *SystemAccrual {
	sem := NewSemaphore(maxSem)
	conn, err := NewConnectionDBAccrual()
	if err != nil {
		logger.Error("Get error while connection to database:", err)
		return nil
	}
	return &SystemAccrual{sem: sem, conn: conn, log: logger}
}

func (s *SystemAccrual) Start() {
	s.log.Infow("Start process accrual system")
	for {
		var wg sync.WaitGroup

		s.log.Infow("Get data...")
		orders, err := s.conn.GetNeedOrders()
		if err != nil {
			s.log.Error("Get error while connection get need orders:", err)
		}
		fmt.Println(orders)
		for idx := 0; idx < len(orders); idx++ {
			wg.Add(1)
			go func(taskID int) {
				s.sem.Acquire()
				defer wg.Done()
				defer s.sem.Release()
				s.log.Info("Run goroutine:", taskID)
				resp, errResp := send(orders[taskID])
				if errResp != nil {
					s.log.Error("Get error while send Get request to system accrual: ", err)
				}
				fmt.Println(resp)
				if errUpd := s.conn.UpdateOrders(resp); errUpd != nil {
					s.log.Error("Get error while update orders: ", errUpd)
				}
				time.Sleep(1 * time.Second)
			}(idx)
		}

		wg.Wait()
		s.log.Infow("All data update")
		time.Sleep(15 * time.Second)
	}
}
