#!/usr/bin/env groovy

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
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    //Note: assetDirectory is on the infrapool agent, not the local Jenkins agent.
  }
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
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }

    stage('Get InfraPool ExecutorV2 Agent(s)') {
      steps{
        script {
          // Request ExecutorV2 agents for 1 hour
          infrapool = getInfraPoolAgent.connected(type: "ExecutorV2", quantity: 1, duration: 1)[0]
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate Changelog and set version') {
      steps {
        script {
          updateVersion(infrapool, "CHANGELOG.md", "${BUILD_NUMBER}")
        }
      }
    }
    stage('Stage running on Atlantis Jenkins Agent Container'){
        steps {
            sh 'scripts/in-container.sh'
        }
    }
    stage('Stage on AWS Instance') {
      steps {
        script {
          // Run script from repo on an AWS instance managed by infrapool
          infrapool.agentSh 'scripts/on-instance.sh'
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
          release(infrapool, { billOfMaterialsDirectory, assetDirectory ->
            /* Publish release artifacts to all the appropriate locations
               Copy any artifacts to assetDirectory on the infrapool node
               to attach them to the Github release.

               If your assets are on the infrapool node in the target
               directory, use a copy like this:
                  infrapool.agentSh "cp target/* ${assetDirectory}"
               Note That this will fail if there are no assets, add :||
               if you want the release to succeed with no assets.

               If your assets are in target on the main Jenkins agent, use:
                 infrapool.agentPut(from: 'target/', to: assetDirectory)
            */
          })
        }
      }
    }
  }
  post {
    always {
      releaseInfraPoolAgent()
    }
  }
}
