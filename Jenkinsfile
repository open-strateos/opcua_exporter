#!groovy

def utils = evaluate readTrusted('./devops/jenkins/utils.groovy')

def DOCKER_TAG = (env.BRANCH_NAME == 'master') ? 'latest' : env.BRANCH_NAME

def OPCUA_DOCKER_IMAGE = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/prometheus/opcua_exporter"
def OPCUA_DIR = "./opcua_exporter"
def SENSORPUSH_DOCKER_IMAGE = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/prometheus/sensorpush_exporter"
def SENSORPUSH_DIR = "./sensorpush_exporter"

pipeline {
  agent { label "infrastructure" }
  environment {
    AWS_DEFAULT_REGION = "us-west-2"
  }
  options {
    skipDefaultCheckout()
  }
  stages {
    stage("Setup") {
      steps {
        script { utils.setup_worker() }
      }
    }

    stage("Test") {
      steps {
        script { utils.test_opcua(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DIR) }
      }
    }

    stage("Build") {
      parallel {
        stage("opcua") {
          steps { script { utils.build_docker(OPCUA_DOCKER_IMAGE, DOCKER_TAG, OPCUA_DIR) } }
        }
        stage("sensorpush") {
          steps { script { utils.build_docker(SENSORPUSH_DOCKER_IMAGE, DOCKER_TAG, SENSORPUSH_DIR) } }
        }
      }
    }

    stage("Push latest") {
      when { branch "master" }
      parallel {
        stage("opcua") {
          steps { sh "docker push ${OPCUA_DOCKER_IMAGE}" }
        }
        stage("sensorpush") {
          steps { sh "docker push ${SENSORPUSH_DOCKER_IMAGE}" }
        }
      }
    }

    stage("Release") {
      when {tag ""} // execute on any tag build
      parallel {
        stage("opcua") {
          steps {
            script { utils.build_docker(OPCUA_DOCKER_IMAGE, TAG_NAME, OPCUA_DIR) }
            sh "docker push ${OPCUA_DOCKER_IMAGE}:${TAG_NAME}" 
          }
        }
        stage("sensorpush") {
          steps {
            script { utils.build_docker(SENSORPUSH_DOCKER_IMAGE, TAG_NAME, SENSORPUSH_DIR) }
            sh "docker push ${SENSORPUSH_DOCKER_IMAGE}:${TAG_NAME}" 
          }
        }
      }
    }

  }

}
