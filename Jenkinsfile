pipeline {
    agent any

    environment {
        goHome = tool 'myGo'
        PATH   = "${goHome}/bin:${env.PATH}"
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
                        script {
                            try {
                                // Build test Docker image
                                echo "Building test Docker image..."
                                sh "docker build -t currency-exchange-test:${env.BUILD_NUMBER} ."
                                
                                // Start container for testing
                                echo "Starting container for integration tests..."
                                sh """
                                docker run -d --name currency-exchange-test-${env.BUILD_NUMBER} \
                                    -p 8080:8080 \
                                    currency-exchange-test:${env.BUILD_NUMBER}
                                """
                                
                                // Wait for service to be ready
                                echo "Waiting for service to be ready..."
                                sh """
                                timeout 60s bash -c 'while ! curl -f http://localhost:8080/health; do sleep 2; done' || {
                                    echo "Service failed to start within 60 seconds"
                                    docker logs currency-exchange-test-${env.BUILD_NUMBER}
                                    exit 1
                                }
                                """
                                
                                // Run integration tests
                                echo "Running integration tests..."
                                sh """
                                export BASE_URL=http://localhost:8080
                                export INTEGRATION=1
                                go test -run TestCurrencyExchangeServiceIntegration -v -timeout 5m
                                """
                                
                            } catch (Exception e) {
                                echo "Integration tests failed: ${e.getMessage()}"
                                throw e
                            } finally {
                                // Always cleanup test container
                                echo "Cleaning up test container..."
                                sh """
                                docker stop currency-exchange-test-${env.BUILD_NUMBER} || true
                                docker rm currency-exchange-test-${env.BUILD_NUMBER} || true
                                docker rmi currency-exchange-test:${env.BUILD_NUMBER} || true
                                """
                            }
                        }
                    },
                    "Coverage": {
                        echo "Running Code Coverage"
                        sh "go test ./internal/service -cover"
                    }
                )
            }
            post {
                always {
                    sh """
                    # Clean up container regardless of test outcome
                    docker stop currency-exchange-test || true
                    docker rm currency-exchange-test || true
                    """
                }
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
