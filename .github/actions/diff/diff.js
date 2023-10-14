function prepareJobs({requirements, configJson}) {
    var map = new Map();
    configJson.forEach(group => {
        console.log(group);
        group.forEach(rule => {
            console.log(rule);
            var name = rule.name;
            var jobName = rule.jobName;
            if (requirements[name] !== "1") {
                return;
            }
            if (map[jobName] !== undefined) {
                map[jobName].push(rule);
            } else {
                map[jobName] = [rule];
            }
        })
    });
    return map;
}

module.exports = {
    prepareJobs
};
