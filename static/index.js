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
    e.preventDefault();

    const dropzone = e.currentTarget;
    dropzone.classList.remove("drop-ready");

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
        this.optionRegex = /option-(\d)/
    }

    connectedCallback() {
        const root = this.attachShadow({ mode: "open" });
        root.innerHTML = `
            <ol id="selected-prios"> </ol>
            <input id="prio-input" list="prio-options" placeholder="Namen der priorisierten Kurse eingeben...">
            <datalist id="prio-options">
             ${this.options.map(name => `
                <option value="${name}">${name}</option>
              `).join('')}
            </datalist>
            <button id="add-prio-button">&oplus;</button>
            `

        this.addSelectedPrio = this.addSelectedPrio.bind(this);
        this.appendHiddenInputs = this.appendHiddenInputs.bind(this);
        this.optionNameToId = this.optionNameToId.bind(this);
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

        if (textInput.value && this.options.includes(textInput.value) && !this.selectedOptions.includes(textInput.value)) {
            const selectedPriosList = this.shadowRoot.querySelector('#selected-prios')
            const li = document.createElement('li');
            li.textContent = textInput.value;
            selectedPriosList.appendChild(li)
        }
    }

    appendHiddenInputs(_) {
            const form = this.closest('form');
            if (!form) return; 

            form.querySelectorAll(`[name="prio[]"]`).forEach(input => input.remove());

            this.selectedOptions.forEach(optName => {
                const hiddenInput = document.createElement('input');
                hiddenInput.type = 'hidden';
                hiddenInput.name = 'prio[]';
                hiddenInput.value = this.optionNameToId(optName);
                form.appendChild(hiddenInput);
            });
    }

    optionNameToId(name) {
        for (const attr of this.attributes) {
            const match = attr.name.match(this.optionRegex)
            if (match) {
                const actualName = attr.value
                if (name === actualName) {
                    return match[1] 
                }
            }
        }
    }

    get selectedOptions() {
        const listItems = this.shadowRoot.querySelectorAll('#selected-prios li');
        if (listItems.length === 0) return [];

        const result = Array.prototype.map.call(listItems, li => li.textContent);

        return result;
    }


    get options() {
        const options = [];

        [...this.attributes].forEach(attr => {
            if (attr.name.startsWith('option-')) {
                options.push(attr.value);
            }
        });

        return options;
    }
}

customElements.define("prio-input", PrioInput)


// I actually want this:
//
// <!DOCTYPE html>
// <html>
// <head>
// <meta name="viewport" content="width=device-width, initial-scale=1">
// <style>
// .focused {
// 	background-color: lightblue;
// }
//
// .show {display: block;}
//
// .chip {
//     display: inline-block;
//     padding: 5px 10px;
//     margin: 5px;
//     background-color: lightgray;
//     border-radius: 16px;
// }
//
// .chip .close {
//     margin-left: 8px;
//     cursor: pointer;
//     color: red;
// }
// </style>
// </head>
// <body style="background-color:white;">
//
// <h2>Search/Filter Dropdown</h2>
// <p>Click on the button to open the dropdown menu, and use the input field to search for a specific dropdown link.</p>
//
// <div class="dropdown">
//   <div id="myDropdown" class="dropdown-content show">
//     <input type="text" placeholder="Search.." id="myInput" onkeyup="filterFunction(event)">
//     <div id="chipContainer"></div>
//     <option value="abc" class="focused">abc</option>
//     <option value="abcd">abcd</option>
//     <option value="abcde">abcde</option>
//     <option value="hjk">hjk</option>
//     <option value="hjkl">hjkl</option>
//   </div>
// </div>
//
// <script>
// let focusedIndex = 0;
//
// function filterFunction(event) {
//   if (event.key === "ArrowUp") {
//      focusedIndex -= 1;
//   } else if (event.key === "ArrowDown") {
//     focusedIndex += 1;
//   } else if (event.key === "Enter") {
//     const selectedOption = document.querySelector('option.focused');
//     if (selectedOption) {
//       addChip(selectedOption.textContent);
//     }
//     return;
//   } else {
//   	focusedIndex = 0;
//   }
//
//   const input = document.getElementById("myInput");
//   const filter = input.value.toUpperCase();
//   const div = document.getElementById("myDropdown");
//   const a = div.getElementsByTagName("option");
//   let showCount = 0;
//   for (let i = 0; i < a.length; i++) {
//     const txtValue = a[i].textContent || a[i].innerText;
//     if (txtValue.toUpperCase().indexOf(filter) > -1) {
//       a[i].style.display = "";
//       showCount += 1;
//
//       if (showCount === focusedIndex) {
//       	a[i].classList.add("focused");
//       } else {
//       	a[i].classList.remove("focused");
//       }
//
//     } else {
//       a[i].style.display = "none";
//     }
//   }
// }
//
// function addChip(text) {
//   const chipContainer = document.getElementById("chipContainer");
//   const chip = document.createElement('span');
//   chip.className = 'chip';
//   chip.textContent = text;
//
//   const closeBtn = document.createElement('span');
//   closeBtn.className = 'close';
//   closeBtn.textContent = 'x';
//   closeBtn.onclick = function() {
//     chipContainer.removeChild(chip);
//   };
//
//   chip.appendChild(closeBtn);
//   chipContainer.appendChild(chip);
// }
// </script>
//
// </body>
// </html>
//
