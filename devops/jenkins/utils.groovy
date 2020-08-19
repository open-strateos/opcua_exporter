#!groovy

/////////////////
// OPCUA Helpers

def test_opcua(image, tag, directory) {

  dir(directory) {
    dockerBuild(
        image: image,
        tag: tag,
        dockerfile: "Dockerfile",
        use_cache: false,
        push: false,
        stage: "tester"
    )
  }
}

def build_docker(image, tag, directory) {

  dir(directory) {
    dockerBuild(
        image: image,
        tag: tag,
        dockerfile: "Dockerfile",
        use_cache: true,
        push: false,
        push_wait: true
    )
  }

}

/////////////////
// System helpers

def setup_worker() {
  checkoutWithChangeSet()
  sh "aws ecr get-login --no-include-email | sh"
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

def cleanup() {
  sh "docker rm -f \$(docker ps -aq) || true"
}

return this
