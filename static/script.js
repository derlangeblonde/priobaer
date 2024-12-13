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
    e.dataTransfer.setData("text", e.target.id)
}

/**
    *@param {DragEvent} e
    */
function drop(e) {
    let participantId = e.dataTransfer.getData("text").split("-")[1]
    let courseId = e.target.id.split("-")[1]

    console.log(participantId)
    console.log(courseId)

    htmx.ajax("PUT", "/assignments", {"target": "#" + e.dataTransfer.getData("text"), "values": {"participant-id": participantId, "course-id": courseId}})
}

/**
    *@param {DragEvent} e
    */
function allowDrop(e) {
    e.preventDefault()
}
