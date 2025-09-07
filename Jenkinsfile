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

        stage("Test") {
            steps {
                parallel(
                    "Unit Test": {
                        echo "=== Running Unit Test ==="
                        sh "go test ./internal/service -v"
                    },
                    "Benchmark Test": {
                        echo "=== Running Benchmark Test ==="
                        sh "go test ./internal/service -bench=. -benchmem"
                    },
                    "Integration Test": {
                        echo "=== Running Integration Test ==="
                        script {
                            try {
                                // Create a Docker network for test isolation
                                echo "Creating test network..."
                                sh "docker network create test-network-${env.BUILD_NUMBER} || true"
                                
                                // Build test Docker image
                                echo "Building test Docker image..."
                                sh "docker build -t currency-exchange-test:${env.BUILD_NUMBER} ."
                                
                                // Start container for testing on the test network
                                echo "Starting container for integration test..."
                                sh """
                                docker run -d --name currency-exchange-test-${env.BUILD_NUMBER} \
                                    --network test-network-${env.BUILD_NUMBER} \
                                    -p 8080:8080 \
                                    currency-exchange-test:${env.BUILD_NUMBER}
                                """
                                
                                // Wait for service to be ready using container network
                                echo "Waiting for service to be ready..."
                                sh """
                                timeout 60s bash -c 'while ! docker run --rm --network test-network-${env.BUILD_NUMBER} curlimages/curl:latest curl -f http://currency-exchange-test-${env.BUILD_NUMBER}:8080/health; do sleep 2; done' || {
                                    echo "Service failed to start within 60 seconds"
                                    docker logs currency-exchange-test-${env.BUILD_NUMBER}
                                    exit 1
                                }
                                """
                                
                                // Run integration test in a container on the same network
                                echo "Running integration test..."
                                sh """
                                docker run --rm --network test-network-${env.BUILD_NUMBER} \
                                    -v \$(pwd):/workspace \
                                    -w /workspace \
                                    -e BASE_URL=http://currency-exchange-test-${env.BUILD_NUMBER}:8080 \
                                    -e INTEGRATION=1 \
                                    golang:1.24-alpine \
                                    sh -c 'go test -run TestCurrencyExchangeServiceIntegration -v -timeout 5m'
                                """
                                
                            } catch (Exception e) {
                                echo "Integration test failed: ${e.getMessage()}"
                                // Show container logs for debugging
                                sh "docker logs currency-exchange-test-${env.BUILD_NUMBER} || true"
                                throw e
                            } finally {
                                // Always cleanup test resources
                                echo "Cleaning up test resources..."
                                sh """
                                docker stop currency-exchange-test-${env.BUILD_NUMBER} || true
                                docker rm currency-exchange-test-${env.BUILD_NUMBER} || true
                                docker network rm test-network-${env.BUILD_NUMBER} || true
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

        // stage("Build Binary") {
        //     steps {
        //         echo "=== Building Go Application ==="
        //         sh "go build -o currency-exchange ./cmd"
        //     }
        // }

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

        stage("Deploy to Kubernetes"){
            steps{
                script{
                    withKubeConfig([credentialsId: 'kubeconfig-staging']){
                        sh "kubectl get pods"
                    }
                }
            }
        }
    }

    post {
        always {
            echo "Pipeline Completed"
        }
        success {
            echo "Pipeline Succeeded"
        }
        failure {
            echo "Pipeline Failed"
        }
    }
}

