const k8s = require('@kubernetes/client-node')
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sClient = kc.makeApiClient(k8s.CustomObjectsApi);

const namespace = process.env.POD_NAMESPACE
const name = process.env.POD_NAME

async function main() {
    try {
        const res = await k8sClient.getNamespacedCustomObject('selenium.mliviusz.com','v1', namespace, 'seleniumtestresults', name);
        console.log(res.body);
        console.log("You have found it, you can update/patch it with new results")
    } catch (e) {
        if(e.statusCode == 404){
            console.log("Not found, so you can create one")
        }
        else {
            console.log(e);
        }
    }
}  

main();

/*var body = {
    "apiVersion": "selenium.mliviusz.com/v1",
    "kind": "SeleniumTestResult",
    "metadata": {
        "name": process.env.POD_NAME,
    },
    "spec": {
        "size": "1",
        "image": "myimage"
    }
}

k8sClient.createNamespacedCustomObject('selenium.mliviusz.com','v1', process.env.POD_NAMESPACE,'seleniumtestresult', body)
    .then((res)=>{
        console.log(res)
    })
    .catch((err)=>{
        console.log(err)
    })
*/
