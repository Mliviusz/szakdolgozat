const runner = require("./runner");

var k8sClient = "";

jest.mock(k8sClient, () => ({
    getNamespacedCustomObject: jest.fn().mockImplementation(arg => ({
        "apiVersion": "selenium.mliviusz.com/v1",
        "kind": "SeleniumTestResult",
        "metadata": {
            "name": "testname"
        },
        "spec": {
            "success": true,
            "endTime": 1678887
        }
    })),
    createNamespacedCustomObject: jest.fn(),
}));

test("Succesfull CreateNewResult", () => {
    expect(runner.createNewResult(true, 1435345)).toBe(0);
});
test("Succesfull UpdateResult", () => {
    expect(runner.updateResult(true, 1435345)).toBe(0);
});