#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { HwTestStack } from '../lib/hw-test-stack';

const app = new cdk.App();
new HwTestStack(app, 'HwTestStack', {
  env: {
    account: process.env.AWS_ACCOUNT_ID,
    region: process.env.AWS_REGION,
  },
});
