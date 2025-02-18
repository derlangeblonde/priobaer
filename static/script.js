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

        this.optionRegex = /^option-(\d+)$/;
         this.appendHiddenInputs = (event) => {
            const form = this.closest('form');
            if (!form) return; 

            form.querySelectorAll(`[name="prio[]"]`).forEach(input => input.remove());

            const selectedPriosList = this.shadowRoot.querySelectorAll('#selected-prios li');
            console.log(selectedPriosList)
            selectedPriosList.forEach(li => {
                const hiddenInput = document.createElement('input');
                hiddenInput.type = 'hidden';
                hiddenInput.name = 'prio[]';
                const courseId = li.attributes['course-id'];
                hiddenInput.value = courseId;
                form.appendChild(hiddenInput);
            });
        };
    }

    connectedCallback() {
        const root = this.attachShadow({ mode: "open" });
        const courseNames = Object.keys(this.optionNamesToId);
        root.innerHTML = `
            <ol id="selected-prios"> </ol>
            <input id="prio-input" list="prio-options" placeholder="Namen der priorisierten Kurse eingeben...">
            <datalist id="prio-options">
             ${courseNames.map(name => `
                <option value="${name}">${name}</option>
              `).join('')}
            </datalist>
            <button id="add-prio-button">&oplus;</button>
            `

        this.addSelectedPrio = this.addSelectedPrio.bind(this);
        root.querySelector('#add-prio-button').addEventListener('click', this.addSelectedPrio)
        const form = this.closest('form');
        if (form) {
            form.addEventListener('submit', this.appendHiddenInputs);
        }

        // TODO: do we need this?
        // htmx.process(root)
    }

    addSelectedPrio(_) {
        const textInput = this.shadowRoot.querySelector('#prio-input');

        if (textInput.value && textInput.value in this.optionNamesToId) {
            const id = this.optionNamesToId[textInput.value];
            const selectedPriosList = this.shadowRoot.querySelector('#selected-prios')
            const li = document.createElement('li');
            li.attributes['course-id'] = id;
            li.textContent = textInput.value;
            selectedPriosList.appendChild(li)
        }
    }

    get optionNamesToId() {
        const options = {};

        [...this.attributes].forEach(attr => {
            const match = attr.name.match(this.optionRegex);
            if (match) {
                const id = parseInt(match[1]);
                options[attr.value] = id;
            }
        });

        return options;
    }
}

customElements.define("prio-input", PrioInput)
