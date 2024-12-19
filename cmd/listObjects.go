package cmd

import (
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"

	"saw/internal"
)

type ListObjectsOpts struct {
	Bucket string

	Prefixes  []string
	Delimiter string
}

var (
	listObjectsOpts = &ListObjectsOpts{
		Bucket:    "",
		Prefixes:  []string{""},
		Delimiter: "",
	}
)

// listObjectsCmd represents the listObjects command
var listObjectsCmd = &cobra.Command{
	Use:   "list-objects",
	Short: "list objects in an s3 bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listObjectsOpts.Bucket == "" {
			return cmd.Help()
		}

		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return err
		}

		client := s3.NewFromConfig(cfg)

		prefixes := internal.SliceCh(listObjectsOpts.Prefixes)

		wg := new(sync.WaitGroup)
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for prefix := range prefixes {
					paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
						Bucket:    &listObjectsOpts.Bucket,
						Prefix:    &prefix,
						Delimiter: &listObjectsOpts.Delimiter,
					})

					for paginator.HasMorePages() {
						output, err := paginator.NextPage(cmd.Context())
						if err != nil {
							slog.Error("failed to list objects", slog.Any("error", err))
							return
						}

						// TODO: print s3 event jsonl
						for _, p := range output.CommonPrefixes {
							cmd.Println(*p.Prefix)
						}
						for _, object := range output.Contents {
							cmd.Println(*object.Key)
						}
					}
				}
			}()
		}
		wg.Wait()
		return nil
	},
}

func init() {
	s3Cmd.AddCommand(listObjectsCmd)

	listObjectsCmd.Flags().StringVarP(&listObjectsOpts.Bucket, "bucket", "b", listObjectsOpts.Bucket, "the bucket to list objects from")
	listObjectsCmd.Flags().StringSliceVarP(&listObjectsOpts.Prefixes, "prefix", "p", listObjectsOpts.Prefixes, "the prefix to filter objects by")
	listObjectsCmd.Flags().StringVarP(&listObjectsOpts.Delimiter, "delimiter", "d", listObjectsOpts.Delimiter, "the delimiter to use when listing objects")
}
