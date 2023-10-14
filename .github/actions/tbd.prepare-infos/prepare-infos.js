const { setClearance } = require('./set_deployment_clearance.js');

function prepareInfos({ process, context, core, workflow_type }) {

    var me_platforms = process.env.ME_PLATFORMS;
    var me_apps = process.env.ME_APPS;
    var orgs = process.env.ORGS;
    var env = process.env.ENV;
    var fe_release_tag = process.env.FE_TAG;
    var be_release_tag = process.env.BE_TAG;
    var me_release_tag = process.env.ME_TAG;
    var auto_deploy = process.env.AUTO_DEPLOY;
    var deploy_all = false;
    var workflow_ref = context.ref;

    var actor = context.actor;
    var alert_channel = "";
    var alert_timestamp = "";

    if (context.payload.client_payload !== undefined) {
        console.log("repository_dispatch", context.payload.client_payload);
        const client_payload = context.payload.client_payload.payload;

        be_release_tag = client_payload.be_release_tag || "";
        fe_release_tag = client_payload.fe_release_tag || "";
        me_release_tag = client_payload.me_release_tag || "";
        env = client_payload.env;
        orgs = client_payload.orgs;
        me_platforms = client_payload.me_platforms;
        me_apps = client_payload.me_apps;
        auto_deploy = client_payload.auto_deploy;
        workflow_ref = client_payload.workflow_ref;

        if (client_payload.slack_alert) {
            actor = client_payload.slack_alert.actor || 'manaops';
            alert_channel = client_payload.slack_alert.channel || "";
            alert_timestamp = client_payload.slack_alert.timestamp || "";
        }
    }

    be_release_tag = be_release_tag.trim();
    fe_release_tag = fe_release_tag.trim();
    me_release_tag = me_release_tag.trim();

    if (be_release_tag && fe_release_tag && me_release_tag) {
        deploy_all = true;
    }

    me_platforms = cleanStringList(me_platforms);
    me_apps = cleanStringList(me_apps);
    orgs = cleanStringList(orgs);

    core.setOutput('actor', actor);
    core.setOutput('alert_channel', alert_channel);
    core.setOutput('alert_timestamp', alert_timestamp);

    core.setOutput('me_platforms', me_platforms);
    core.setOutput('me_apps', me_apps);
    core.setOutput('orgs', orgs);
    core.setOutput('env', env);
    core.setOutput('be_release_tag', be_release_tag);
    core.setOutput('fe_release_tag', fe_release_tag);
    core.setOutput('me_release_tag', me_release_tag);
    core.setOutput('auto_deploy', auto_deploy);
    core.setOutput('deploy_all', deploy_all);


    core.setOutput('workflow_ref', be_release_tag || workflow_ref);

    // Deploy for JPREP after deploying for Manabie, instead of at the same time,
    // so that each case has more cpu to work with.
    core.setOutput('max_deploy_parallel', 1);


    setClearance({
        core: core,
        workflow_type: workflow_type,
        environment: env,
        organizations: orgs,
        be_release_tag,
        fe_release_tag,
        me_release_tag,
        me_apps,
        me_platforms,
    });
}

function cleanStringList(input = '') {
    const array = input.split(",").map(s => s.trim());
    return array.join(', ');
}

module.exports = {
    prepareInfos
};
