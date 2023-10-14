module.exports = async (tag) => {
    checkValidTag(tag);
    if (tag.includes('rc')) {
        return 'staging';
    } else {
        return 'uat';
    }
}

function checkValidTag(tag) {
    const isValid = tag.match(/v\d+.\d+(-rc\d+)*/);
    if (!isValid) {
       // temporary disabled for hotfix
       // throw 'The tag is invalid'; 
    }
}
