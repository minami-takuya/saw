package cmd

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/spf13/cobra"

	"github.com/minamitakuya/saw/internal"
)

type SendMessageOpts struct {
	QueueURL    string
	Input       string
	MessageSize int
}

var (
	sendMessageOpts = &SendMessageOpts{
		QueueURL:    "",
		Input:       "",
		MessageSize: 10,
	}
)

// sendMessageCmd represents the sendMessage command
var sendMessageCmd = &cobra.Command{
	Use:   "send-message",
	Short: "send a message to an sqs queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sendMessageOpts.QueueURL == "" {
			return cmd.Help()
		}
		slog.Info("sending message", slog.Any("queue-url", sendMessageOpts.QueueURL), slog.Any("input", sendMessageOpts.Input))

		var in io.ReadCloser

		in = os.Stdin
		if sendMessageOpts.Input != "" {
			f, err := os.Open(sendMessageOpts.Input)
			if err != nil {
				return err
			}
			in = f
		}
		defer in.Close()

		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return err
		}
		client := sqs.NewFromConfig(cfg)

		lines := internal.ReadByLine(in)
		chunkedLines := internal.ChunkCh(lines, sendMessageOpts.MessageSize)

		wg := new(sync.WaitGroup)
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for chunk := range chunkedLines {
					entries := make([]types.SendMessageBatchRequestEntry, 0, sendMessageOpts.MessageSize)
					for i, line := range chunk {
						id := strconv.Itoa(i)
						entries = append(entries, types.SendMessageBatchRequestEntry{
							Id:          &id,
							MessageBody: &line,
						})
					}

					_, err := client.SendMessageBatch(cmd.Context(), &sqs.SendMessageBatchInput{
						QueueUrl: &sendMessageOpts.QueueURL,
						Entries:  entries,
					})

					if err != nil {
						slog.Error("failed to send message", slog.Any("error", err))
					}
				}
			}()
		}

		wg.Wait()
		return nil
	},
}

func init() {
	sqsCmd.AddCommand(sendMessageCmd)

	sendMessageCmd.Flags().StringVarP(&sendMessageOpts.QueueURL, "queue-url", "q", sendMessageOpts.QueueURL, "the url of the sqs queue")
	sendMessageCmd.Flags().StringVarP(&sendMessageOpts.Input, "input", "i", sendMessageOpts.Input, "the message to send")
	sendMessageCmd.Flags().IntVarP(&sendMessageOpts.MessageSize, "message-size", "m", sendMessageOpts.MessageSize, "the size of the message to send")
}
