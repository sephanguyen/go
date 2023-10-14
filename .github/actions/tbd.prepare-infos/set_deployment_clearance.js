function setClearance({
    core,
    workflow_type,
    environment,
    organizations,
    be_release_tag,
    fe_release_tag,
    me_release_tag,
    me_apps,
    me_platforms,
}) {

    if (workflow_type != 'build' && workflow_type != 'deploy') {
        console.log('Invalid workflow type: ' + workflow_type);
        return;
    }

    const filtered_organizations = filterOrganizationsInput(workflow_type, environment, organizations);

    const jobClearance = getJobsClearance({
        be_release_tag,
        fe_release_tag,
        me_release_tag,
        me_apps,
        me_platforms,
        filtered_organizations,
        environment
    })

    console.log("jobClearance", JSON.stringify(jobClearance, null, 4))

    core.setOutput('backend_k8s_orgs', jobClearance.backend_k8s_orgs);
    core.setOutput('backoffice_k8s_orgs', jobClearance.backoffice_k8s_orgs);
    core.setOutput('teacher_k8s_orgs', jobClearance.teacher_k8s_orgs);
    core.setOutput('learner_k8s_orgs', jobClearance.learner_k8s_orgs);
    core.setOutput('backoffice_firebase_orgs', jobClearance.backoffice_firebase_orgs);
    core.setOutput('learner_firebase_orgs', jobClearance.learner_firebase_orgs);
    core.setOutput('learner_android_orgs', jobClearance.learner_android_orgs);
    core.setOutput('learner_ios_orgs', jobClearance.learner_ios_orgs);
    core.setOutput('teacher_firebase_orgs', jobClearance.teacher_firebase_orgs);
    core.setOutput('appsmith_orgs', jobClearance.appsmith_orgs);
}

const buildConditions = {
    'production': {
        'backend': {
            'k8s': ['tokyo', 'jprep', 'synersia', 'renseikai', 'ga', 'aic']
        },
        'backoffice': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo', 'jprep']
        },
        'learner': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo', 'jprep'],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep']
        },
        'teacher': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo', 'jprep']
        }
    },

    'preproduction': {
        'backend': {
            'k8s': ['tokyo']
        },
        'backoffice': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo']
        },
        'learner': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo'],
            'android': [],
            'ios': []
        },
        'teacher': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['tokyo']
        }
    },

    'staging': {
        'backend': {
            'k8s': ['manabie', 'jprep']
        },
        'backoffice': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
        },
        'learner': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep']
        },
        'teacher': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
        }
    },

    'uat': {
        'backend': {
            'k8s': ['manabie', 'jprep']
        },
        'backoffice': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
        },
        'learner': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep']
        },
        'teacher': {
            'firebase': [], // it will re-use build from k8s jobs
            'k8s': ['manabie', 'jprep'],
        }
    },
}

// On PRODUCTION, 'tokyo' and 'manabie' is the same org but:
// - deploy-backend uses 'tokyo'
// - deploy-app uses 'manabie'
// - deploy-app iOS only uses 'manabie' and 'jprep'
// - deploy-web only uses 'jprep', 'synersia', 'renseikai', 'ga'

const deployConditions = {
    'production': {
        'backend': {
            'k8s': ['tokyo', 'jprep', 'synersia', 'renseikai', 'ga', 'aic']
        },
        'backoffice': {
            'firebase': ['jprep'],
            'k8s': ['tokyo', 'jprep'],
        },
        'appsmith': {
            'host': ['tokyo', 'synersia', 'renseikai', 'ga', 'aic'],
        },
        'learner': {
            'firebase': ['jprep'],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep'],
            'k8s': ['tokyo', 'jprep'],
        },
        'teacher': {
            'firebase': ['jprep'],
            'k8s': ['tokyo', 'jprep'],
        }
    },

    'preproduction': {
        'backend': {
            'k8s': ['tokyo']
        },
        'backoffice': {
            'firebase': [],
            'k8s': ['tokyo']
        },
        'learner': {
            'firebase': [],
            'k8s': ['tokyo'],
            'android': [],
            'ios': []
        },
        'teacher': {
            'firebase': [],
            'k8s': ['tokyo'],
        }
    },

    'staging': {
        'backend': {
            'k8s': ['manabie', 'jprep']
        },
        'backoffice': {
            'k8s': ['manabie', 'jprep'],
            'firebase': ['jprep']
        },
        'appsmith': {
            'host': ['manabie'],
        },
        'learner': {
            'k8s': ['manabie', 'jprep'],
            'firebase': [],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep']
        },
        'teacher': {
            'k8s': ['manabie', 'jprep'],
            'firebase': []
        }
    },

    'uat': {
        'backend': {
            'k8s': ['manabie', 'jprep']
        },
        'backoffice': {
            'k8s': ['manabie', 'jprep'],
            'firebase': []
        },
        'appsmith': {
            'host': ['manabie'],
        },
        'learner': {
            'firebase': [],
            'android': ['manabie', 'jprep'],
            'ios': ['manabie', 'jprep'],
            'k8s': ['manabie', 'jprep'],
        },
        'teacher': {
            'firebase': [],
            'k8s': ['manabie', 'jprep'],
        }
    },
}

function filterOrganizationsInput(workflow_type, environment, organizations) {

    const inputOrgs = typeof organizations === "string" ? organizations.split(",").map(s => s.trim()) : [...organizations];
    const organizationsList = ['manabie', 'jprep', 'tokyo', 'synersia', 'renseikai', 'ga', 'aic']

    console.log("inputOrgs", inputOrgs)
    const validOrgs = inputOrgs.filter(org => organizationsList.includes(org));
    if (!validOrgs.length) {
        console.log('All organizations provided are invalid: ' + inputOrgs);
        return;
    }


    const condition = workflow_type === "deploy" ? deployConditions : buildConditions;
    const env = condition[environment];
    if (!env) {
        console.log("Invalid environment provided: " + environment);
        return;
    }



    // Since these orgs: aic, ga, renseikai, synersia will be using prod-tokyo database,
    // we must deploy for tokyo first, then these orgs later.
    // See https://manabie.atlassian.net/browse/LT-29678
    const backend_k8s_orgs = filterOrg(validOrgs, env.backend.k8s).sort((a, b) => {
        if (b === "tokyo") return -1;
        return 0;
    }).reverse();

    const backoffice_k8s_orgs = filterOrg(validOrgs, env.backoffice.k8s);
    const teacher_k8s_orgs = filterOrg(validOrgs, env.teacher.k8s);
    const learner_k8s_orgs = filterOrg(validOrgs, env.learner.k8s);
    const backoffice_firebase_orgs = filterOrg(validOrgs, env.backoffice.firebase);
    const learner_firebase_orgs = filterOrg(validOrgs, env.learner.firebase);
    const learner_android_orgs = filterOrg(validOrgs, env.learner.android);
    const learner_ios_orgs = filterOrg(validOrgs, env.learner.ios);
    const teacher_firebase_orgs = filterOrg(validOrgs, env.learner.firebase);
    const appsmith_orgs = filterOrg(validOrgs, env.appsmith?.host);



    return {
        backend_k8s_orgs,
        backoffice_k8s_orgs,
        teacher_k8s_orgs,
        learner_k8s_orgs,
        backoffice_firebase_orgs,
        learner_firebase_orgs,
        learner_android_orgs,
        learner_ios_orgs,
        teacher_firebase_orgs,
        appsmith_orgs
    }
}

function getJobsClearance({
    be_release_tag,
    fe_release_tag,
    me_release_tag,
    me_apps,
    me_platforms,
    filtered_organizations,
    environment
}) {

    if (!filtered_organizations) {
        const clearance = {
            backend_k8s_orgs: [],
            backoffice_k8s_orgs: [],
            teacher_k8s_orgs: [],
            learner_k8s_orgs: [],
            backoffice_firebase_orgs: [],
            learner_firebase_orgs: [],
            learner_android_orgs: [],
            learner_ios_orgs: [],
            teacher_firebase_orgs: [],
            appsmith_orgs: []
        }

        console.log("clearance", JSON.stringify(clearance, null, 2))
        return clearance;
    };

    let {
        backend_k8s_orgs,
        backoffice_k8s_orgs,
        backoffice_firebase_orgs,
        teacher_k8s_orgs,
        teacher_firebase_orgs,
        learner_k8s_orgs,
        learner_firebase_orgs,
        learner_android_orgs,
        learner_ios_orgs,
        appsmith_orgs
    } = filtered_organizations;

    const is_deploy_teacher_web = me_release_tag && me_apps.includes('teacher');
    const is_deploy_learner_web = me_release_tag && me_apps.includes('learner') && me_platforms.includes('web');
    const is_deploy_learner_android = me_release_tag && me_apps.includes('learner') && me_platforms.includes('android');
    const is_deploy_learner_ios = me_release_tag && me_apps.includes('learner') && me_platforms.includes('ios');
    const is_deploy_backend_k8s = Boolean(be_release_tag);

    console.log({
        is_deploy_teacher_web,
        is_deploy_learner_web,
        is_deploy_learner_android,
        is_deploy_learner_ios,
        is_deploy_backend_k8s
    })


    const clearance = {
        // k8s part
        backend_k8s_orgs: is_deploy_backend_k8s ? backend_k8s_orgs : [],
        backoffice_k8s_orgs: fe_release_tag ? backoffice_k8s_orgs : [],
        teacher_k8s_orgs: is_deploy_teacher_web ? teacher_k8s_orgs : [],
        learner_k8s_orgs: is_deploy_learner_web ? learner_k8s_orgs : [],
        // firebase part
        backoffice_firebase_orgs: fe_release_tag ? backoffice_firebase_orgs : [],
        teacher_firebase_orgs: is_deploy_teacher_web ? teacher_firebase_orgs : [],
        learner_firebase_orgs: is_deploy_learner_web ? learner_firebase_orgs : [],
        // android part
        learner_android_orgs: is_deploy_learner_android ? learner_android_orgs : [],
        // ios part
        learner_ios_orgs: is_deploy_learner_ios ? learner_ios_orgs : [],
        // appsmith part
        appsmith_orgs: fe_release_tag ? appsmith_orgs : []
    }

    console.log("clearance", JSON.stringify(clearance, null, 2))

    return clearance;
}

function filterOrg(validOrgs = [], orgs = []) {
    return orgs.filter(org => validOrgs.includes(org)) || [];
}

module.exports = {
    setClearance
};
