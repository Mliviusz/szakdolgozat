const k8s = require('@kubernetes/client-node')
const execSync = require('child_process').execSync;
const fs = require('fs');
const path = require('path')
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sClient = kc.makeApiClient(k8s.CustomObjectsApi);

const namespace = process.env.POD_NAMESPACE
const name = process.env.POD_NAME

async function main() {
    var success = false;
    var endTime = -1;
    try{
        console.log("Starting Selenium tests");
        execSync('selenium-side-runner -s $SELENIUM_GRID -o results -r $RETRIES /mnt/config/*.side');
        const jsonsInDir = fs.readdirSync('./results').filter(file => path.extname(file) === '.json');
        jsonsInDir.forEach(file => { // There will be only 1 file but with timestamp in name
            const fileData = fs.readFileSync(path.join('./results', file));
            const json = JSON.parse(fileData.toString());
            success = json.success;
            endTime = json.testResults[0].endTime
          });
    } catch (e) {
        console.log(e);
    }
    try {
        const res = await k8sClient.getNamespacedCustomObject('selenium.mliviusz.com','v1', namespace, 'seleniumtestresults', name);
        console.log("Updating SeleniumTestResult")
        updateResult(success, endTime)
    } catch (e) {
        if(e.statusCode == 404){
            console.log("Creating SeleniumTestResult")
            createNewResult(success, endTime);
        }
        else {
            console.log(e);
        }
    }
}

async function updateResult(success, endTime) {
    const patch = [{
        "op": "replace",
        "path":"/spec/endTime", 
        "value": endTime
    },
    {
        "op": "replace",
        "path":"/spec/success", 
        "value": success
    }];
    const options = { "headers": { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH}};

    k8sClient.patchNamespacedCustomObject ('selenium.mliviusz.com','v1', namespace,'seleniumtestresults', name, patch, undefined, undefined, undefined, options)
    .then((res)=>{
        console.log("SeleniumTestResult " + name + " updated")
    })
    .catch((err)=>{
        console.log(err)
        return 1;
    })
    return 0;
}

async function createNewResult(success, endTime) {
    var body = {
        "apiVersion": "selenium.mliviusz.com/v1",
        "kind": "SeleniumTestResult",
        "metadata": {
            "name": name
        },
        "spec": {
            "success": success,
            "endTime": endTime
        }
    }

    k8sClient.createNamespacedCustomObject('selenium.mliviusz.com','v1', namespace, 'seleniumtestresults', body)
    .then((res)=>{
        console.log("SeleniumTestResult " + name + " created")
    })
    .catch((err)=>{
        console.log(err)
        return 1;
    })

    return 0;
}

main();