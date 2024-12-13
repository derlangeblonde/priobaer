window.addEventListener("beforeunload", function (event) {

    const navigationType = performance.getEntriesByType("navigation")[0]?.type;

    if (navigationType === "navigate" || navigationType === "back_forward") {
        event.preventDefault();
        event.returnValue = '';
    }
});

/**
    *@param {DragEvent} e
    */
function dragStart(e) {
    e.dataTransfer.setData("css-id", e.target.id)
}

/**
    *@param {DragEvent} e
    */
function allowDrop(e) {
    e.preventDefault()
}

/**
    *@param {DragEvent} e
    */
function drop(e) {
    let participantElementId = e.dataTransfer.getData("css-id")
    let participantId = extractNumericId(participantElementId)
    let courseId = extractNumericId(e.target.id) 

    htmx.ajax("PUT", "/assignments", {"target": "#" + participantElementId, "values": {"participant-id": participantId, "course-id": courseId}})
}

function extractNumericId(elementId) {
    return elementId.split("-")[1]
}
