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
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Gather the SUMS file to sign
    sums = "*SHA256SUMS"
    sh "cp ${assetDirectory}/${sums} ./"
    sumFile = sh(script: "ls ${sums}", returnStdout: true).trim()
    // Create .tar for signing
    sh "mv ${sumFile} ${sumFile}.tar"
    signArtifacts patterns: ["${sumFile}.tar"]
    // Rename the sig file
    sh "mv ${sumFile}.tar.sig ${sumFile}.sig"
    // Copy back to assetDirectory
    sigLocation = pwd() + "/${sumFile}.sig"
    sh "cp ${sigLocation} ${assetDirectory}"
  }
  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-AmznDocker' }

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

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Build artifacts') {
      steps {
        script {
          sh './bin/build'
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
            sh './bin/test'
          }
        }
      }
      post {
        always {
          withCredentials([
            conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_username", variable: 'INFRAPOOL_SHARED_SERVICES_TENANT'),
            conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_address", variable: 'INFRAPOOL_SHARED_SERVICES_DOMAIN'),
            conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_password", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET'),
            conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_username", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_ID'),
          ]){
            script {
              sh './bin/cleanup-resources.sh'
              normalizeOwnership('output')
              stash name: 'output-xml', includes: 'output/*.xml'
            }
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
            sh './bin/integration-test.sh'
          }
        }
      }
      post {
        always {
          script {
            withCredentials([
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_username", variable: 'INFRAPOOL_SHARED_SERVICES_TENANT'),
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-tenant_address", variable: 'INFRAPOOL_SHARED_SERVICES_DOMAIN'),
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_password", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET'),
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_sh-shared-services-client_username", variable: 'INFRAPOOL_SHARED_SERVICES_CLIENT_ID'),
            ]){
              script {
                sh './bin/cleanup-resources.sh'
              }
            }
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
          release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations
            // Copy any artifacts to assetDirectory to attach them to the Github release
            sh "cp -r dist/*.zip dist/*_SHA256SUMS ${assetDirectory}"
            // Create Go module SBOM
            sh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          }
        }
      }
    }
  }

  post {
    always {
      unstash 'output-xml'
      junit 'output/junit.xml'
      recordCoverage(tools: [[parser: 'COBERTURA', pattern: 'output/coverage.xml']])
      codacy action: 'reportCoverage', filePath: "output/coverage.xml"
    }
  }
}

// Fixes root ownership left behind by Docker volume mounts so that subsequent
// Jenkins steps and re-runs can read and write the affected directories.
void normalizeOwnership(String... dirs) {
  sh """
    if ! command -v sudo >/dev/null 2>&1; then
      exit 0
    fi
    u="\$(id -u):\$(id -g)"
    for d in ${dirs.join(' ')}; do
      if [ -e "\$d" ]; then
        sudo chown -R "\$u" "\$d" 2>/dev/null || true
        sudo chmod -R u+rwX "\$d" 2>/dev/null || true
      fi
    done
  """
}