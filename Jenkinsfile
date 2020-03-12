#!groovy

def DOCKER_TAG = (env.BRANCH_NAME == 'master') ? 'latest' : env.BRANCH_NAME

def OPCUA_DOCKER_IMAGE = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/prometheus/opcua_exporter"
def OPCUA_DOCKERFILE = "./opcua_exporter/Dockerfile"

timeout(time: 40, unit: 'MINUTES') {
  node('docker') {
    withEnv([
        'AWS_DEFAULT_REGION=us-west-2'
    ]) {
      try {
        stage('Setup') {
          checkout scm
          sh "aws ecr get-login --no-include-email | sh"
        }

        stage('Test') {
          parallel(
              "opcua": { test_opcua(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DOCKERFILE) }
          )
        }

        stage('Build Artifact') {
          parallel(
              "opcua": { build_opcua(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DOCKERFILE) }
          )
        }

      } catch (err) {
        notify_failure()
        throw err
      } finally {
        sh "docker rm -f \$(docker ps -aq) || true"
      }
    }
  }
}

def test_opcua(image, tag, dockerfile) {

  dockerBuild(
      image: image,
      tag: tag,
      dockerfile: dockerfile,
      use_cache: true,
      push: false,
      stage: "tester"
  )
}
def build_opcua(image, tag, dockerfile) {

  dockerBuild(
      image: image,
      tag: tag,
      dockerfile: dockerfile,
      use_cache: true,
      push: true,
      push_wait: true
  )
}

def notify_failure() {
  if (env.BRANCH_NAME == 'master') {
    slackSend(
        channel: '#devops',
        color: 'danger',
        message: ":samjoch_ohlala: ${env.JOB_NAME} failure! #${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
    )
  }
}
