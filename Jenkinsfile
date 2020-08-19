pipeline {
    agent any

    environment {
        IMAGE_NAME = "panacea-core"
        GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
    }

    stages {
        stage('Start') {
            steps {
                slackSend (channel: '#alerts-ci', color: '#FFFF00', message: "STARTED: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
            }
        }
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'docker build -t ${IMAGE_NAME}:${GIT_COMMIT_HASH} .'
            }
        }
        stage('Test') {
            steps {
                echo 'Testing..'
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploying..'
            }
        }
    }

    post {
        always {
            sh 'docker rmi ${IMAGE_NAME}:${GIT_COMMIT_HASH} || true'
        }
        success {
            slackSend (channel: '#alerts-ci', color: '#00FF00', message: "SUCCESSFUL: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
        }
        failure {
            slackSend (channel: '#alerts-ci', color: '#FF0000', message: "FAILED: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
        }
    }
}