// PaletteEditor renders and manages a compact palette table UI

const HEX_REGEX = /^#[0-9a-fA-F]{6}$/;

export class PaletteEditor {
    constructor({ allowAdd = true } = {}) {
        this.allowAdd = allowAdd;
        this.container = null;
        this.rowsContainer = null;
    }

    render(initialValues = []) {
        this.container = document.createElement('div');
        this.container.className = 'palette-colors-field';
        this.container.setAttribute('data-field-type', 'palette_colors');

        const header = document.createElement('div');
        header.className = 'palette-colors-header';
        ['Use', 'Color', 'Hex', ''].forEach((text) => {
            const cell = document.createElement('div');
            cell.textContent = text;
            header.appendChild(cell);
        });
        this.container.appendChild(header);

        this.rowsContainer = document.createElement('div');
        this.rowsContainer.className = 'palette-colors-rows';
        this.container.appendChild(this.rowsContainer);

        const addBtn = document.createElement('button');
        addBtn.type = 'button';
        addBtn.className = 'btn';
        addBtn.textContent = 'Add color';
        addBtn.style.display = this.allowAdd ? 'inline-flex' : 'none';
        addBtn.addEventListener('click', () => this.addRow('#000000', true));
        this.container.appendChild(addBtn);

        // Populate initial rows
        if (Array.isArray(initialValues) && initialValues.length) {
            initialValues.forEach((color) => {
                const enabled = !color.startsWith('!');
                const value = enabled ? color : color.slice(1);
                this.addRow(value || '#000000', enabled);
            });
        } else {
            this.addRow('#000000', true);
        }

        return this.container;
    }

    addRow(colorValue = '#000000', enabled = true) {
        if (!this.rowsContainer) return;

        const row = document.createElement('div');
        row.className = 'palette-color-row';

        const includeCheckbox = document.createElement('input');
        includeCheckbox.type = 'checkbox';
        includeCheckbox.checked = enabled;
        includeCheckbox.title = 'Include this color in the palette';
        includeCheckbox.addEventListener('change', () => {
            row.classList.toggle('disabled', !includeCheckbox.checked);
        });

        const colorInput = document.createElement('input');
        colorInput.type = 'color';
        colorInput.className = 'form-input color-picker';
        colorInput.value = HEX_REGEX.test(colorValue) ? colorValue : '#000000';

        const textInput = document.createElement('input');
        textInput.type = 'text';
        textInput.className = 'form-input color-text';
        textInput.value = colorInput.value;
        textInput.placeholder = '#RRGGBB';

        const removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'btn palette-remove-btn';
        removeBtn.textContent = '✕';
        removeBtn.addEventListener('click', () => {
            row.remove();
        });

        colorInput.addEventListener('input', () => {
            textInput.value = colorInput.value;
        });

        textInput.addEventListener('input', () => {
            if (HEX_REGEX.test(textInput.value)) {
                colorInput.value = textInput.value;
            }
        });

        row.appendChild(includeCheckbox);
        row.appendChild(colorInput);
        row.appendChild(textInput);
        row.appendChild(removeBtn);
        row.classList.toggle('disabled', !enabled);

        this.rowsContainer.appendChild(row);
    }

    getValues() {
        if (!this.rowsContainer) return [];
        return Array.from(this.rowsContainer.querySelectorAll('.palette-color-row')).map((row) => {
            const include = row.querySelector('input[type="checkbox"]')?.checked ?? true;
            const val = row.querySelector('.color-text')?.value.trim() ?? '';
            if (!val) return '';
            return include ? val : `!${val}`;
        }).filter(Boolean);
    }

    validate(fieldName = 'colors') {
        const errors = [];
        const values = this.getValues();
        const container = this.container;

        const invalid = values.filter((v) => {
            const raw = v.startsWith('!') ? v.slice(1) : v;
            return !HEX_REGEX.test(raw);
        });

        if (invalid.length) {
            errors.push(`Field "${fieldName}" has invalid color(s): ${invalid.join(', ')}`);
            container?.classList.add('error');
        } else {
            container?.classList.remove('error');
        }

        return { valid: errors.length === 0, errors };
    }
}
