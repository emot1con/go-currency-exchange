pipeline {
    agent any

    environment {
        goHome = tool 'myGo'
        PATH   = "${goHome}/bin:${env.PATH}"
    NO_PROXY = '127.0.0.1,localhost'
    }

    stages {
        stage("Checkout") {
            steps {
                echo "=== Checkout Source Code ==="
                sh "go version"
                sh "echo PATH - $PATH"
                sh "echo BUILD_NUMBER - $env.BUILD_NUMBER"
            }
        }

        stage("Tests") {
            steps {
                parallel(
                    "Unit Tests": {
                        echo "=== Running Unit Tests ==="
                        sh "go test ./internal/service -v"
                    },
                    "Benchmark Tests": {
                        echo "=== Running Benchmark Tests ==="
                        sh "go test ./internal/service -bench=. -benchmem"
                    },
                                        "Integration Tests": {
                        echo "=== Running Integration Tests ==="
                        sh """
                                                set -euxo pipefail
                                                docker rm -f currency-exchange-test || true
                                                docker run -d --name currency-exchange-test -p 8080:8080 numpyh/currency-exchange:jenkins-test-go-pipeline-21
                                                # Wait for readiness via health endpoint
                                                for i in $(seq 1 30); do
                                                    if curl -fsS http://127.0.0.1:8080/health | grep -qi healthy; then
                                                        break
                                                    fi
                                                    sleep 1
                                                done
                                                export BASE_URL=http://127.0.0.1:8080
                                                export INTEGRATION=1
                                                go test -run TestIntegrationOnly -v || { echo "Integration tests failed. Container logs:"; docker logs currency-exchange-test || true; exit 1; }
                                                docker rm -f currency-exchange-test || true
                        """
                    },
                    "Coverage": {
                        echo "Running Code Coverage"
                        sh "go test ./internal/service -cover"
                    }
                )
            }
        }

        stage("Build Binary") {
            steps {
                echo "=== Building Go Application ==="
                sh "go build -o currency-exchange ./cmd"
            }
        }

        stage("Build Docker Image") {
            steps {
                script {
                    dockerImage = docker.build("numpyh/currency-exchange:${env.BUILD_TAG}")
                }
            }
        }

        stage("Push Docker Image") {
            steps {
                script {
                    docker.withRegistry('', 'dockerhub') {
                        dockerImage.push()
                        dockerImage.push('latest')
                    }
                }
            }
        }
    }
}
