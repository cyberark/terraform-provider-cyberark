#!/usr/bin/env groovy

@Library("product-pipelines-shared-library") _

  // Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { infrapool, sourceVersion, targetVersion, assetDirectory ->

  }
  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  environment {
    MODE = release.canonicalizeMode()
  }

  stages {
    stage('Scan for internal URLs') {
      steps {
        script {
          detectInternalUrls()
        }
      }
    }

    stage('Get InfraPool ExecutorV2 Agent') {
      steps {
        script {
          // Request InfraPool
          INFRAPOOL_EXECUTORV2_AGENT_0 = getInfraPoolAgent.connected(type: "ExecutorV2", quantity: 1, duration: 1)[0]
        }
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
          // Copy the vendor directory onto infrapool
          INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "vendor", to: "${WORKSPACE}"
          INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "go.*", to: "${WORKSPACE}"
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion(INFRAPOOL_EXECUTORV2_AGENT_0, "CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Build artifacts') {
      steps {
        script {
          INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/build'
        }
      }
    }

    stage('Test') {
      steps {
        withCredentials([
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_username", variable: 'INFRAPOOL_SHARED_SERVICES_TENANT'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_address", variable: 'INFRAPOOL_SHARED_SERVICES_DOMAIN'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_password", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_username", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_ID'),

          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_username", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_NAME'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_address", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_ALIAS'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_password", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_REGION'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-account_username", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_ACCOUNT_ID'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-account_password", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_IAM_ROLE')
        ]){
          script {
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/test'
          }
        }
      }
      post {
        always {
          script {
            INFRAPOOL_EXECUTORV2_AGENT_0.agentStash name: 'output-xml', includes: 'output/*.xml'
          }
        }
      }
    }

    stage('Run integration tests for P-Cloud and SecretsHub') {
      steps {
        withCredentials([
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_username", variable: 'INFRAPOOL_SHARED_SERVICES_TENANT'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_address", variable: 'INFRAPOOL_SHARED_SERVICES_DOMAIN'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_password", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_username", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_ID'),

          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_username", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_NAME'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_address", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_ALIAS'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-nar_password", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_REGION'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-account_username", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_ACCOUNT_ID'),
          conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-aws-account_password", variable: 'INFRAPOOL_SHARED_SERVICES_AWS_IAM_ROLE')
        ]){
          script {
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/integration-test.sh'
          }
        }
      }
    }
    
    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          release(INFRAPOOL_EXECUTORV2_AGENT_0) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations
            // Copy any artifacts to assetDirectory to attach them to the Github release
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "cp -r dist/*.zip dist/*_SHA256SUMS ${assetDirectory}"
            // Create Go module SBOM
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          }
        }
      }  
    }
  }
  post {
    always {
      unstash 'output-xml'
      junit 'output/junit.xml'
      cobertura autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: 'output/coverage.xml', conditionalCoverageTargets: '30, 0, 0', failUnhealthy: false, failUnstable: false, lineCoverageTargets: '30, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '30, 0, 0', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
      codacy action: 'reportCoverage', filePath: "output/coverage.xml"

      releaseInfraPoolAgent(".infrapool/release_agents")
    }
  }
}
