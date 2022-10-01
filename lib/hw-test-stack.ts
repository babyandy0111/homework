import * as cdk from '@aws-cdk/core';
import * as sqs from '@aws-cdk/aws-sqs';
import * as path from 'path';
import * as lambda from '@aws-cdk/aws-lambda';
import * as lambdaEventSources from '@aws-cdk/aws-lambda-event-sources';
import * as apigw from '@aws-cdk/aws-apigateway';

export class HwTestStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const stageName = 'dev'
    const userEndpointURL = 'user'
    const friendEndpointURL = 'friend'
    const subscriptionEndpointURL = 'subscription'
    const DBNAME = process.env.DB_NAME || ''
    const DBHOST = process.env.DB_HOST || ''
    const DBPW = process.env.DB_PW || ''
    const DBUSER = process.env.DB_USER || ''

    const queue1 = new sqs.Queue(this, 'HwTestUserQueue', {
      queueName: 'user',
      visibilityTimeout: cdk.Duration.seconds(300),
    })

    const queue2 = new sqs.Queue(this, 'HwTestFriendQueue', {
      queueName: 'friend',
      visibilityTimeout: cdk.Duration.seconds(300),
    })

    const queue3 = new sqs.Queue(this, 'HwTestSBQueue', {
      queueName: 'subscription',
      visibilityTimeout: cdk.Duration.seconds(300),
    })

    // 讓lambda 訂閱 queue
    const userLambda = new lambda.Function(this, 'user', {
      runtime: lambda.Runtime.GO_1_X,
      timeout: cdk.Duration.seconds(300),
      handler: "main",
      code: lambda.Code.fromAsset(
        path.join(__dirname, "../lambda/user")
      ),
      environment: {
        region: cdk.Stack.of(this).region,
        zones: JSON.stringify(cdk.Stack.of(this).availabilityZones),
        db_name: DBNAME,
        db_host: DBHOST,
        db_pw: DBPW,
        db_user: DBUSER,
      },
    })

    const eventSource1 = new lambdaEventSources.SqsEventSource(queue1);

    userLambda.addEventSource(eventSource1);


    const friendLambda = new lambda.Function(this, 'friend', {
      runtime: lambda.Runtime.GO_1_X,
      timeout: cdk.Duration.seconds(300),
      handler: "main",
      code: lambda.Code.fromAsset(
        path.join(__dirname, "../lambda/friend")
      ),
      environment: {
        region: cdk.Stack.of(this).region,
        zones: JSON.stringify(cdk.Stack.of(this).availabilityZones),
        db_name: DBNAME,
        db_host: DBHOST,
        db_pw: DBPW,
        db_user: DBUSER,
      },
    })

    const eventSource2 = new lambdaEventSources.SqsEventSource(queue2);

    friendLambda.addEventSource(eventSource2);


    const subscriptionLambda = new lambda.Function(this, 'subscription', {
      runtime: lambda.Runtime.GO_1_X,
      timeout: cdk.Duration.seconds(300),
      handler: "main",
      code: lambda.Code.fromAsset(
        path.join(__dirname, "../lambda/subscription")
      ),
      environment: {
        region: cdk.Stack.of(this).region,
        zones: JSON.stringify(cdk.Stack.of(this).availabilityZones),
        db_name: DBNAME,
        db_host: DBHOST,
        db_pw: DBPW,
        db_user: DBUSER,
      },
    })

    const eventSource3 = new lambdaEventSources.SqsEventSource(queue3);

    subscriptionLambda.addEventSource(eventSource3);

    // 新增一個apige ~ 非必須
    const restApi = new apigw.RestApi(this, "dev-api", { deploy: false });

    restApi.root
      .addResource(userEndpointURL)
      .addMethod("GET", new apigw.LambdaIntegration(userLambda, { proxy: true }));

    restApi.root
      .addResource(friendEndpointURL)
      .addMethod("GET", new apigw.LambdaIntegration(friendLambda, { proxy: true }));

    restApi.root
      .addResource(subscriptionEndpointURL)
      .addMethod("GET", new apigw.LambdaIntegration(subscriptionLambda, { proxy: true }));

    const devDeploy = new apigw.Deployment(this, "dev-deployment", { api: restApi });

    // 定義stage => 網址會是 /dev/bot
    new apigw.Stage(this, "dev-stage", {
      deployment: devDeploy,
      stageName: stageName
    });

    // print出位置
    new cdk.CfnOutput(this, "user URL", {
      value: `https://${restApi.restApiId}.execute-api.${this.region}.amazonaws.com/${stageName}/${userEndpointURL}`,
    });
    new cdk.CfnOutput(this, "friend URL", {
      value: `https://${restApi.restApiId}.execute-api.${this.region}.amazonaws.com/${stageName}/${friendEndpointURL}`,
    });
    new cdk.CfnOutput(this, "subscription URL", {
      value: `https://${restApi.restApiId}.execute-api.${this.region}.amazonaws.com/${stageName}/${subscriptionEndpointURL}`,
    });
  }
}
