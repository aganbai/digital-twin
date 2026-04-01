#!/bin/bash
cd /Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend
npx jest --config e2e/jest.config.js e2e/full-flow.test.js --forceExit --testTimeout=120000 > /tmp/e2e-result.log 2>&1
echo "EXIT_CODE=$?" >> /tmp/e2e-result.log
touch /tmp/e2e-done.flag
