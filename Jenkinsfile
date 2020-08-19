pipeline {
    agent any

    stages {
        stage('Start') {
            steps {
                slackSend (channel: '#alerts-ci', color: '#FFFF00', message: "STARTED: ${BUILD_TAG} (${env.BUILD_URL})")
            }
        }
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'make build'
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
        success {
            slackSend (channel: '#alerts-ci', color: '#00FF00', message: "SUCCESSFUL: ${BUILD_TAG} (${env.BUILD_URL})")
        }
        failure {
            slackSend (channel: '#alerts-ci', color: '#FF0000', message: "FAILED: ${BUILD_TAG} (${env.BUILD_URL})")
        }
    }
}