package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/slack-go/slack"
)

var SLACK_BOT_TOKEN string = "xoxb-2793245968321-2780648689043-R9PqXdmLKUPs0s2hIhgN4NRg"
var api = slack.New(SLACK_BOT_TOKEN)
var regionList []string = []string{"ap-northeast-2", "ap-southeast-2", "ap-southeast-1", "ap-northeast-1", "us-east-1", "us-west-2"}

func GetRunningInstances(client *ec2.EC2) {
	result, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Couldn't retrieve running instances: %v", err)
		sendSlack(fmt.Sprintf("Unable to list instances, %v", err))

		return
	}
	sendSlack(fmt.Sprintln("=====================EC2====================="))
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			sendSlack(fmt.Sprintf("Found running EC2 instance: %s\n", *instance.InstanceId))
		}
	}

}

func GetRdsInstances(rdsclient *rds.RDS) {
	sendSlack(fmt.Sprintln("=====================RDS====================="))
	result, err := rdsclient.DescribeDBInstances(nil)
	if err != nil {
		exitErrorf("Unable to list instances, %v", err)
		sendSlack(fmt.Sprintf("Unable to list instances, %v", err))

	}
	for _, d := range result.DBInstances {
		// fmt.Printf("* %s created on %s\n", aws.StringValue(d.DBInstanceIdentifier), aws.TimeValue(d.InstanceCreateTime))
		sendSlack(fmt.Sprintf("Found running RDS EndPoint: %s\n", *d.Endpoint.Address))
	}
}

func GetDynamoInstances(dynamoclient *dynamodb.DynamoDB) {
	limit := new(int64)
	*limit = int64(10)
	tableList, err := dynamoclient.ListTables(&dynamodb.ListTablesInput{
		Limit: limit,
	})
	if err != nil {
		exitErrorf("Unable to list instances, %v", err)
		sendSlack(fmt.Sprintf("Unable to list instances, %v", err))
	}

	sendSlack(fmt.Sprintln("===================DynamoDB=================="))
	for _, n := range tableList.TableNames {
		sendSlack(fmt.Sprintf("Found running DynamoDB Table: %s\n", *n))
	}

}

func sendSlack(str string) {
	_, _, err := api.PostMessage(
		"SlackNumber",
		slack.MsgOptionText(str, false),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func main() {
	t := time.Now()
	for _, val := range regionList {

		sess, err := session.NewSessionWithOptions(session.Options{
			Profile: "default",
			Config: aws.Config{
				Region: aws.String(val),
			},
		})

		if err != nil {
			log.Printf("Failed to initialize new session: %v", err)
			sendSlack(fmt.Sprintf("Failed to initialize new session: %v", err))
			return
		}

		ec2client := ec2.New(sess)
		rdsclient := rds.New(sess)
		dynamoclient := dynamodb.New(sess)

		sendSlack(fmt.Sprintf("%d년%d월%d일  %d시%d분\t%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), val))
		GetRunningInstances(ec2client)

		GetRdsInstances(rdsclient)
		GetDynamoInstances(dynamoclient)

		sendSlack(fmt.Sprintln("=============================================="))
		sendSlack(fmt.Sprintln("Message sent successfully"))
		fmt.Println("Message sent successfully")
	}
	fmt.Println("aws-service-List-slack success")
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
