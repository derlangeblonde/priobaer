// deno-lint-ignore-file no-unused-vars
window.addEventListener("beforeunload", function (event) {
    const navigationType = performance.getEntriesByType("navigation")[0]?.type;

    if (navigationType === "navigate" || navigationType === "back_forward") {
        event.preventDefault();
        event.returnValue = "";
    }
});

/**
 * @param {DragEvent} e
 */
function dragStart(e) {
    e.dataTransfer.setData("css-id", e.target.id);
}

/**
 * @param {DragEvent} e
 */
function allowDrop(e) {
    e.preventDefault();
    e.target.classList.add("drop-ready")
}

function dragLeave(e) {
    e.preventDefault();
    e.target.classList.remove("drop-ready")
}

/**
 * @param {DragEvent} e
 */
function drop(e) {
    const participantElementId = e.dataTransfer.getData("css-id");
    const participantId = extractNumericId(participantElementId);
    const courseElementId = e.target.id;

    const payload = {
        "participant-id": participantId,
    };

    if (courseElementId !== "not-assigned") {
        payload["course-id"] = extractNumericId(courseElementId);
    }

    htmx.ajax("PUT", "/assignments", {
        "target": "#" + participantElementId,
        "values": payload,
    });
}

/**
 * @param {string} elementId
 */
function extractNumericId(elementId) {
    return elementId.split("-")[1];
}

document.addEventListener('DOMContentLoaded', function() {
    const orderedList = document.getElementById('ordered-list');

    // Function to add selected options to the ordered list
    window.addSelectedOptions = function() {
        const select = document.getElementById('options');
        const selectedOptions = Array.from(select.selectedOptions);

        selectedOptions.forEach(option => {
            const listItem = document.createElement('li');
            listItem.textContent = option.textContent;
            listItem.setAttribute('data-value', option.value);
            orderedList.appendChild(listItem);
        });

        // Clear selected options from the select box
        selectedOptions.forEach(option => option.selected = false);
    };

    // Make the ordered list sortable
    // new Sortable(orderedList, {
    //     animation: 150
    // });

    // Handle form submission
    document.getElementById('ordered-list-form').addEventListener('submit', function(event) {
        event.preventDefault();
        const orderedValues = Array.from(orderedList.children).map(li => li.getAttribute('data-value'));
        htmx.ajax("POST", "/prio", {
            "target": "body",
            "values": orderedValues,
        });
        // You can submit the orderedValues to the server here
    });
});

// Sortable library (you can include this from a CDN or install via npm)
/*!
 * Sortable 1.14.0
 * Released under the MIT license
 * https://github.com/SortableJS/Sortable
 */
// (function (root, factory) {
//     if (typeof define === 'function' && define.amd) {
//         define(factory);
//     } else if (typeof exports === 'object') {
//         module.exports = factory();
//     } else {
//         root.Sortable = factory();
//     }
// }(this, function () {
//     'use strict';
//
//     // ... (Sortable library code)
//
//     return Sortable;
// }));

