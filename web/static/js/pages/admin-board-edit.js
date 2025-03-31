/* web/static/js/pages/admin-board-edit.js */
document.addEventListener('alpine:init', () => {
    Alpine.data('boardEditForm', () => ({
        submitting: false,
        fields: [],
        fieldCount: 0,
        nextId: -1,

        init() {
            // 스크립트 태그에서 필드 데이터 초기화
            try {
                const initialFieldsScript = document.getElementById('initial-field-data');
                if (initialFieldsScript && initialFieldsScript.textContent) {
                    // initialFieldsScript.textContent가 이미 문자열이므로 직접 확인
                    let parsedFields;
                    const content = initialFieldsScript.textContent.trim();

                    // 이미 JSON 문자열인지 확인 (따옴표로 시작하는지)
                    if (content.startsWith('"') && content.endsWith('"')) {
                        // 이미 문자열화된 JSON인 경우 이스케이프된 따옴표 처리
                        const unescapedContent = content.slice(1, -1).replace(/\\"/g, '"');
                        parsedFields = JSON.parse(unescapedContent);
                    } else {
                        // 일반 JSON인 경우
                        parsedFields = JSON.parse(content);
                    }

                    // 배열이 아닌 경우 배열로 변환
                    if (!Array.isArray(parsedFields)) {
                        if (typeof parsedFields === 'object' && parsedFields !== null) {
                            parsedFields = [parsedFields]; // 객체인 경우 배열로 변환
                        } else {
                            parsedFields = []; // 그 외의 경우 빈 배열로
                        }
                    }

                    // 기존 필드에 isNew 속성 추가
                    this.fields = parsedFields.map(field => ({
                        ...field,
                        isNew: false,
                        // columnName이 없는 경우 name으로 설정
                        columnName: field.columnName || field.name
                    }));
                } else {
                    this.fields = [];
                }
            } catch (e) {
                this.fields = [];
            }

            // 다음 ID 설정 (새 필드용)
            this.nextId = -1;

            // fields 배열의 변경을 감시
            this.$watch('fields', value => {
                this.fieldCount = value.length;
            });
        },

        // 필드 추가 메소드
        addField() {
            this.fields.push({
                id: this.nextId--,
                isNew: true,
                name: '',
                columnName: '', // 명시적으로 columnName 속성 추가
                displayName: '',
                fieldType: 'text',
                required: false,
                sortable: false,
                searchable: false,
                options: ''
            });
        },

        // 필드 제거 메소드
        removeField(index) {
            this.fields.splice(index, 1);
        },

        // 폼 제출 메소드
        submitForm() {
            this.submitting = true;

            // 폼 요소 가져오기
            const form = document.getElementById('board-edit-form');

            // FormData 객체 생성
            const formData = new FormData(form);
            formData.append('field_count', this.fields.length);

            // 필드 데이터 디버깅용 객체
            const debugFields = {};

            // 각 필드에 대한 더 자세한 정보 수집
            const fieldsDetails = this.fields.map((field, index) => {
                return {
                    index,
                    id: field.id,
                    name: field.name,
                    columnName: field.columnName || field.name,
                    displayName: field.displayName,
                    isNew: field.isNew,
                    fieldType: field.fieldType
                };
            });

            this.fields.forEach((field, index) => {
                // 컬럼명 결정 - 기존 필드는 columnName을, 새 필드는 name을 사용
                const columnName = field.isNew ? field.name : (field.columnName || field.name);

                debugFields[`field_${index}`] = {
                    id: field.id,
                    name: field.name,
                    columnName: columnName,
                    isNew: field.isNew
                };

                // 폼 데이터에 필드 정보 명시적으로 추가
                formData.set(`field_id_${index}`, field.id);
                formData.set(`field_name_${index}`, columnName); // 중요: 이 부분을 columnName으로 설정
                formData.set(`display_name_${index}`, field.displayName);
                formData.set(`field_type_${index}`, field.fieldType);

                // 체크박스는 체크된 경우에만 값이 전송되므로 명시적으로 설정
                formData.set(`required_${index}`, field.required ? "on" : "off");
                formData.set(`sortable_${index}`, field.sortable ? "on" : "off");
                formData.set(`searchable_${index}`, field.searchable ? "on" : "off");

                // select 필드의 경우 옵션 추가
                if (field.fieldType === 'select' && field.options) {
                    formData.set(`options_${index}`, field.options);
                }
            });

            // FormData의 모든 키-값 쌍을 로깅
            let formEntries = {};
            for (let [key, value] of formData.entries()) {
                formEntries[key] = value;
            }

            // 폼 액션 URL 가져오기
            const actionUrl = form.getAttribute('action');

            // CSRF 토큰 가져오기
            const csrfToken = formData.get('csrf_token');

            // 서버에 데이터 전송
            fetch(actionUrl, {
                method: 'PUT',
                headers: {
                    'X-CSRF-Token': csrfToken,
                    'Accept': 'application/json'
                },
                body: formData
            })
                .then(res => this.handleResponse(res))
                .catch(err => this.handleError(err));
        },

        // 응답 처리 메소드
        handleResponse(res) {
            // Content-Type 확인
            const contentType = res.headers.get('Content-Type');

            // JSON 응답인 경우
            if (contentType && contentType.includes('application/json')) {
                return res.json().then(data => {
                    if (data.success) {
                        window.location.href = '/admin/boards';
                    } else {
                        alert(data.message);
                        this.submitting = false;
                    }
                });
            }
            // HTML 응답 (오류 페이지)인 경우
            else {
                return res.text().then(html => {
                    alert('처리 중 오류가 발생했습니다. 다시 시도해 주세요.');
                    this.submitting = false;
                });
            }
        },

        // 오류 처리 메소드
        handleError(err) {
            alert('오류가 발생했습니다: ' + err);
            this.submitting = false;
        }
    }));
});