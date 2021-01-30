package qsortm

type task struct {
	startPos, endPos int
}

func channelInverter(inputCh, outputCh chan task) {
	taskBuffer := []task{}
	for {
		newTask, ok := <-inputCh
		if !ok {
			// the input channel has already closed
			break
		}
		taskBuffer = append(taskBuffer, newTask)
		for len(taskBuffer) > 0 {
			select {
			case outputCh <- taskBuffer[len(taskBuffer)-1]:
				taskBuffer = taskBuffer[:len(taskBuffer)-1]
			case newTask := <-inputCh:
				taskBuffer = append(taskBuffer, newTask)
			}
		}
	}

	close(outputCh)
}
