const path = require("path");
const fs = require("fs");

const { setClearance } = require('../tbd.prepare-infos/set_deployment_clearance.js');

const EXCEPTIONS = ["root"];
function setupMfeJob({
    rootPath,
    core,
    context,
    inputs
}) {
    let {release_tag, build_and_deploy_root_only, squads, fragments, orgs, env, workflow_type, auto_deploy} = inputs;


    const squadPath = path.resolve(rootPath, "src/squads");

    const dir = fs.readdirSync(squadPath);
    const dirFiltered = dir.filter((folder)=> {
        return !EXCEPTIONS.includes(folder)
    })

    let teams = {};

    dirFiltered.forEach((name)=> {
        const teamPath = path.resolve(squadPath, name);
        const manifestFileName = `${env}-${name}-manifest.json`
        const manifestPath = path.resolve(teamPath, manifestFileName);

        teams[name] = {
            teamName: name,
            teamPath: teamPath.replace(rootPath, "./"),
            manifestPath: manifestPath.replace(rootPath, "./"),
            manifestFileName: manifestFileName,
        }
    })
    const rootShell = path.join(squadPath, "root").replace(rootPath, "./");

    console.log("rootShell", rootShell)
    console.log("teams", JSON.stringify(teams, null, 2))

    core.setOutput("mfe-root", rootShell);
    core.setOutput("mfe-teams", Object.values(teams)); 

    const teamNames = Object.keys(teams);
    core.setOutput("mfe-team-infos", JSON.stringify(teams));
    core.setOutput("mfe-team-names", teamNames);

    // set inputs
    let context_ref =  context.ref;

    if (context && context.payload && context.payload.client_payload) {
        const payload = context.payload.client_payload.payload
        console.log("payload", JSON.stringify(payload, null, 2))

        build_and_deploy_root_only = payload.build_and_deploy_root_only;
        context_ref = payload.context_ref;
        release_tag = payload.release_tag;
        squads = payload.squads;
        fragments = payload.fragments;
        orgs = payload.orgs;
        env = payload.env;
        auto_deploy = payload.auto_deploy;
        workflow_type = payload.workflow_type;

    }

    const orgArrays = orgs?.split(",").map(org => org.trim()).filter(org => org) || []
    const fragmentArrays = fragments ? fragments?.split(",").map(fragment => fragment.trim()).filter(fragment => fragment) : [];
    const squadArrays = squads ? squads?.split(",").map(squad => squad.trim()).filter(squad => squad) : teamNames;

    const fragmentStringArgs = !fragmentArrays.length ? '' : `--fragments ${fragmentArrays.join(" --fragments ")}`
    const teamStringArgs = !squadArrays.length ? '' : `--teams ${squadArrays.join(" --teams ")}`

    console.log({
        release_tag,
        build_and_deploy_root_only,
        context_ref,
        squadArrays,
        teamStringArgs,
        fragmentArrays,
        fragmentStringArgs,
        orgArrays,
        env,
    })
    
    core.setOutput("build_and_deploy_root_only", `${build_and_deploy_root_only}` === "true" ? "true" : "");
    core.setOutput("release_tag", release_tag);
    core.setOutput("context_ref", context_ref);


    core.setOutput("squads", squadArrays);
    core.setOutput("mfe-teams-args", teamStringArgs)
    core.setOutput("orgs", orgArrays); // will use orgs from setClearance fn
    core.setOutput("fragments", fragmentArrays);
    core.setOutput("mfe-fragments-args", fragmentStringArgs);
    core.setOutput("env", env);
    core.setOutput("auto_deploy", auto_deploy);


    setClearance({ 
        core,
        workflow_type: workflow_type,
        environment: env,
        organizations: orgArrays,
        be_release_tag: "",
        fe_release_tag: release_tag,
        me_release_tag: "",
        me_apps: [],
        me_platforms: [],
    })
}

module.exports = {
    setupMfeJob
}
