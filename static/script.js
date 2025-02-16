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

class PrioInput extends HTMLElement {
    constructor() {
        super();
        console.log("PrioInput constructor");
    }

    connectedCallback() {
        console.log("connected Callback")
        const root = this.attachShadow({ mode: "open" });
        const currentOptions = this.options
        root.innerHTML = `
            <ol id="selected-prios">
            </ol>
            <input id="prio-input" list="prio-options" placeholder="Namen der priorisierten Kurse eingeben...">
            <datalist id="prio-options">
             ${currentOptions.map(opt => `
                <option value="${opt}">${opt}</option>
              `).join('')}
            </datalist>
            <button id="add-prio-button">&oplus;</button>
            `
        console.log(root.innerHTML)

        this.selectPrio = this.selectPrio.bind(this);
        root.querySelector('#add-prio-button').addEventListener('click', this.selectPrio)

        // TODO: do we need this?
        // htmx.process(root)
    }

    selectPrio(_) {
        const textInput = this.shadowRoot.querySelector('#prio-input');

        if (textInput.value && this.options.includes(textInput.value)) {
            const selectedPriosList = this.shadowRoot.querySelector('#selected-prios')
            const li = document.createElement('li');
            li.textContent = textInput.value;
            selectedPriosList.appendChild(li)
        }
    }

    get options() {
        const options = [];

        [...this.attributes].forEach(attr => {
            if (attr.name.includes('option')) {
            options.push(attr.value);
            }
        });

        return options;
    }
}

customElements.define("prio-input", PrioInput)
