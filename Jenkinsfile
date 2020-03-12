#!groovy

def utils = evaluate readTrusted('./devops/jenkins/utils.groovy')

def DOCKER_TAG = (env.BRANCH_NAME == 'master') ? 'latest' : env.BRANCH_NAME

def OPCUA_DOCKER_IMAGE = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/prometheus/opcua_exporter"
def OPCUA_DIR = "./opcua_exporter"

timeout(time: 10, unit: 'MINUTES') {
  node('docker') {
    withEnv([
        'AWS_DEFAULT_REGION=us-west-2'
    ]) {
      try {
        stage('Setup') {
          utils.setup_worker()
        }

        stage('Test') {
          parallel(
              "opcua": { utils.test_opcua(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DIR) }
          )
        }

        stage('Build Artifact') {
          parallel(
              "opcua": { utils.build_opcua(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DIR) }
          )
        }

      } catch (err) {
        utils.notify_failure()
        throw err
      } finally {
        utils.cleanup()
      }
    }
  }
}
