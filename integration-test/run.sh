#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
set -x

if [[ -z "${FURYA_DAPP_REPO:-}" ]]; then
    commit=7e968801a0a03f47f59dd7683f1653935222ea88
    rm -fr furya-dapp
    git clone https://github.com/FURYA/furya-dapp.git
    cd furya-dapp
    git checkout $commit
else
    cd $FURYA_DAPP_REPO
fi

yarn

npx tsx packages/scripts/integration-testing/simpleTest ..
npx tsx packages/scripts/integration-testing/upgradeTest ..