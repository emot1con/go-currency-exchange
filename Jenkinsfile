pipeline{
    agent any

    environment{
        myGo = tool 'myGo'
        PATH = "${myGo}/bin:${env.PATH}"
    }

    stages{
        stage("Checkout"){
            steps{
                echo "Checkout Source Code"
                echo "PATH: ${env.PATH}"
                echo "Go Version: ${sh(script: 'go version', returnStdout: true).trim()}"
                echo "BUIILD_NUMBER: ${env.BUILD_NUMBER}"
            }
        }

        stage("Test"){
            steps{

            parallel{
                "Unit Test"{
                    echo "Unit Test"
                    sh "go test ./internal/service -v"
                },
                "Benchmark Test"{
                    echo "Benchmark Test"
                    sh "go test -bench=. ./internal/service -v"
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
                "Coverage Test"{
                    echo "Coverage Test"
                    sh "go test ./internal/service -cover "
                }
            }
            }
        }

        stage("Build Docker Image"){
            steps{
                echo "Build Docker Image"
                dockerImage = docker.build("numpy/currency-exchange:${env.BUILD_NUMBER}")
            }
        }

        stage("Push Docker Image"){
            steps{
                echo "Push Docker Image"
                docker.withRegistry("", "dockerhub"){
                    dockerImage.push()
                    dockerImage.push('latest')
                }
            }
        }
    }

    post{
        always{
            echo "Pipeline Completed"
        }
        success{
            echo "Pipeline Succeeded"
        }
        failure{
            echo "Pipeline Failed"
        }
    }
}