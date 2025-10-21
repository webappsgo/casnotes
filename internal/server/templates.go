package server

// HTML Templates per CLAUDE.md User Interface
// Dracula-based dark theme as default

const baseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Title}} - casnotes</title>
	<style>
		:root {
			/* Dracula theme colors per CLAUDE.md */
			--bg: #282a36;
			--current-line: #44475a;
			--selection: #44475a;
			--foreground: #f8f8f2;
			--comment: #6272a4;
			--cyan: #8be9fd;
			--green: #50fa7b;
			--orange: #ffb86c;
			--pink: #ff79c6;
			--purple: #bd93f9;
			--red: #ff5555;
			--yellow: #f1fa8c;
		}

		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
			background: var(--bg);
			color: var(--foreground);
			line-height: 1.6;
		}

		.container {
			max-width: 1400px;
			margin: 0 auto;
			padding: 20px;
		}

		header {
			background: var(--current-line);
			padding: 1rem 2rem;
			box-shadow: 0 2px 10px rgba(0,0,0,0.3);
			position: sticky;
			top: 0;
			z-index: 1000;
		}

		header .header-content {
			max-width: 1400px;
			margin: 0 auto;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}

		header h1 {
			color: var(--green);
			font-size: 1.5rem;
		}

		header nav a {
			color: var(--cyan);
			text-decoration: none;
			margin-left: 2rem;
			transition: color 0.3s;
		}

		header nav a:hover {
			color: var(--green);
		}

		.btn {
			display: inline-block;
			padding: 0.5rem 1rem;
			border-radius: 5px;
			text-decoration: none;
			font-weight: 500;
			transition: all 0.3s;
			border: none;
			cursor: pointer;
		}

		.btn-primary {
			background: var(--green);
			color: var(--bg);
		}

		.btn-primary:hover {
			background: var(--cyan);
		}

		.btn-secondary {
			background: var(--purple);
			color: var(--foreground);
		}

		.btn-secondary:hover {
			background: var(--pink);
		}

		.btn-danger {
			background: var(--red);
			color: var(--foreground);
		}

		.btn-danger:hover {
			background: var(--orange);
		}

		footer {
			margin-top: 4rem;
			padding: 2rem;
			text-align: center;
			color: var(--comment);
			border-top: 1px solid var(--current-line);
		}

		.alert {
			padding: 1rem;
			margin: 1rem 0;
			border-radius: 5px;
		}

		.alert-success {
			background: var(--green);
			color: var(--bg);
		}

		.alert-error {
			background: var(--red);
			color: var(--foreground);
		}

		.alert-warning {
			background: var(--orange);
			color: var(--bg);
		}

		.alert-info {
			background: var(--cyan);
			color: var(--bg);
		}

		/* Light theme per CLAUDE.md (GitHub-inspired) */
		body.light {
			--bg: #ffffff;
			--current-line: #f6f8fa;
			--selection: #d1e4f8;
			--foreground: #24292e;
			--comment: #6a737d;
			--cyan: #0366d6;
			--green: #28a745;
			--orange: #f66a0a;
			--pink: #e36209;
			--purple: #6f42c1;
			--red: #d73a49;
			--yellow: #ffd33d;
		}
	</style>
	{{.ExtraCSS}}
</head>
<body class="{{.Theme}}">
	<header>
		<div class="header-content">
			<h1>📝 casnotes</h1>
			<nav>
				{{if .User}}
					<a href="/users">Dashboard</a>
					<a href="/users/notes">Notes</a>
					<a href="/users/settings">Settings</a>
					{{if .User.IsAdmin}}
						<a href="/admin">Admin</a>
					{{end}}
					<a href="/logout">Logout</a>
				{{else}}
					<a href="/">Home</a>
					<a href="/discover">Discover</a>
					<a href="/login">Login</a>
					<a href="/register">Register</a>
				{{end}}
			</nav>
		</div>
	</header>

	<main class="container">
		{{.Content}}
	</main>

	<footer>
		<p>&copy; 2024 casnotes - Self-hosted notes | <a href="/privacy" style="color: var(--cyan);">Privacy</a> | <a href="/terms" style="color: var(--cyan);">Terms</a> | <a href="/.well-known/security.txt" style="color: var(--cyan);">Security</a></p>
	</footer>

	{{.ExtraJS}}
</body>
</html>`

// Grid view per CLAUDE.md (Google Keep style)
const gridViewTemplate = `<style>
	.view-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding: 1rem;
		background: var(--current-line);
		border-radius: 5px;
	}

	.view-toggle {
		display: flex;
		gap: 0.5rem;
	}

	.view-toggle button {
		padding: 0.5rem 1rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
		cursor: pointer;
		transition: all 0.3s;
	}

	.view-toggle button.active {
		background: var(--green);
		color: var(--bg);
		border-color: var(--green);
	}

	.search-bar {
		flex: 1;
		max-width: 500px;
		margin: 0 2rem;
	}

	.search-bar input {
		width: 100%;
		padding: 0.5rem 1rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
		font-size: 1rem;
	}

	.notes-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.note-card {
		background: var(--current-line);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.3s;
		cursor: pointer;
		position: relative;
		overflow: hidden;
	}

	.note-card:hover {
		transform: translateY(-5px);
		box-shadow: 0 5px 15px rgba(0,0,0,0.3);
	}

	.note-card.pinned {
		border: 2px solid var(--yellow);
	}

	.note-card.archived {
		opacity: 0.6;
	}

	.note-card .note-title {
		font-size: 1.1rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
		color: var(--cyan);
	}

	.note-card .note-content {
		color: var(--foreground);
		font-size: 0.9rem;
		line-height: 1.4;
		max-height: 150px;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.note-card .note-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}

	.note-tag {
		background: var(--purple);
		color: var(--foreground);
		padding: 0.2rem 0.5rem;
		border-radius: 3px;
		font-size: 0.75rem;
	}

	.note-card .note-meta {
		margin-top: 0.5rem;
		font-size: 0.75rem;
		color: var(--comment);
		display: flex;
		justify-content: space-between;
	}

	.fab {
		position: fixed;
		bottom: 2rem;
		right: 2rem;
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: var(--green);
		color: var(--bg);
		font-size: 2rem;
		border: none;
		cursor: pointer;
		box-shadow: 0 4px 12px rgba(0,0,0,0.4);
		transition: all 0.3s;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.fab:hover {
		background: var(--cyan);
		transform: scale(1.1);
	}
</style>

<div class="view-controls">
	<div class="view-toggle">
		<button class="active" onclick="setView('grid')">Grid</button>
		<button onclick="setView('list')">List</button>
		<button onclick="setView('timeline')">Timeline</button>
	</div>
	<div class="search-bar">
		<input type="text" placeholder="Search notes..." id="searchInput">
	</div>
	<a href="/users/notes/new" class="btn btn-primary">+ New Note</a>
</div>

<div class="notes-grid">
	{{range .Notes}}
	<div class="note-card {{if .Pinned}}pinned{{end}} {{if .Archived}}archived{{end}}" onclick="location.href='/users/notes/{{.ID}}'">
		{{if .Pinned}}<div style="position: absolute; top: 10px; right: 10px;">📌</div>{{end}}
		<div class="note-title">{{.Title}}</div>
		<div class="note-content">{{.Content}}</div>
		{{if .Tags}}
		<div class="note-tags">
			{{range .Tags}}
			<span class="note-tag">{{.}}</span>
			{{end}}
		</div>
		{{end}}
		<div class="note-meta">
			<span>{{.NoteType}}</span>
			<span>{{.UpdatedAt}}</span>
		</div>
	</div>
	{{end}}
</div>

<button class="fab" onclick="location.href='/users/notes/new'">+</button>

<script>
	function setView(view) {
		location.href = '/users/notes?view=' + view;
	}

	document.getElementById('searchInput').addEventListener('input', function(e) {
		const query = e.target.value.toLowerCase();
		const cards = document.querySelectorAll('.note-card');
		cards.forEach(card => {
			const title = card.querySelector('.note-title').textContent.toLowerCase();
			const content = card.querySelector('.note-content').textContent.toLowerCase();
			if (title.includes(query) || content.includes(query)) {
				card.style.display = 'block';
			} else {
				card.style.display = 'none';
			}
		});
	});
</script>`

// List view per CLAUDE.md
const listViewTemplate = `<style>
	.notes-list {
		background: var(--current-line);
		border-radius: 8px;
		overflow: hidden;
	}

	.note-row {
		display: grid;
		grid-template-columns: auto 1fr auto auto;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid var(--bg);
		cursor: pointer;
		transition: background 0.3s;
	}

	.note-row:hover {
		background: var(--selection);
	}

	.note-row:last-child {
		border-bottom: none;
	}

	.note-row .note-icon {
		font-size: 1.5rem;
		display: flex;
		align-items: center;
	}

	.note-row .note-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.note-row .note-title {
		font-weight: 600;
		color: var(--cyan);
	}

	.note-row .note-snippet {
		color: var(--comment);
		font-size: 0.85rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 600px;
	}

	.note-row .note-date {
		color: var(--comment);
		font-size: 0.85rem;
		white-space: nowrap;
	}

	.note-row .note-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.note-row .note-actions button {
		padding: 0.25rem 0.5rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 3px;
		cursor: pointer;
		font-size: 0.75rem;
	}

	.note-row .note-actions button:hover {
		background: var(--purple);
		border-color: var(--purple);
	}
</style>

<div class="view-controls">
	<div class="view-toggle">
		<button onclick="setView('grid')">Grid</button>
		<button class="active" onclick="setView('list')">List</button>
		<button onclick="setView('timeline')">Timeline</button>
	</div>
	<div class="search-bar">
		<input type="text" placeholder="Search notes..." id="searchInput">
	</div>
	<a href="/users/notes/new" class="btn btn-primary">+ New Note</a>
</div>

<div class="notes-list">
	{{range .Notes}}
	<div class="note-row" onclick="location.href='/users/notes/{{.ID}}'">
		<div class="note-icon">
			{{if eq .NoteType "note"}}📝{{end}}
			{{if eq .NoteType "code"}}💻{{end}}
			{{if eq .NoteType "checklist"}}✅{{end}}
			{{if eq .NoteType "canvas"}}🎨{{end}}
			{{if eq .NoteType "encrypted"}}🔒{{end}}
			{{if .Pinned}}📌{{end}}
		</div>
		<div class="note-info">
			<div class="note-title">{{.Title}}</div>
			<div class="note-snippet">{{.Content}}</div>
		</div>
		<div class="note-date">{{.UpdatedAt}}</div>
		<div class="note-actions" onclick="event.stopPropagation()">
			<button onclick="pinNote('{{.ID}}')">Pin</button>
			<button onclick="archiveNote('{{.ID}}')">Archive</button>
		</div>
	</div>
	{{end}}
</div>

<button class="fab" onclick="location.href='/users/notes/new'">+</button>

<script>
	function setView(view) {
		location.href = '/users/notes?view=' + view;
	}

	function pinNote(id) {
		fetch('/api/v1/notes/' + id + '/pin', {method: 'POST'})
			.then(() => location.reload());
	}

	function archiveNote(id) {
		fetch('/api/v1/notes/' + id + '/archive', {method: 'POST'})
			.then(() => location.reload());
	}
</script>`

// Timeline view per CLAUDE.md
const timelineViewTemplate = `<style>
	.timeline {
		position: relative;
		padding-left: 3rem;
	}

	.timeline::before {
		content: '';
		position: absolute;
		left: 1rem;
		top: 0;
		bottom: 0;
		width: 2px;
		background: var(--current-line);
	}

	.timeline-item {
		position: relative;
		margin-bottom: 2rem;
	}

	.timeline-item::before {
		content: '';
		position: absolute;
		left: -2.5rem;
		top: 0.5rem;
		width: 1rem;
		height: 1rem;
		border-radius: 50%;
		background: var(--green);
		border: 2px solid var(--bg);
	}

	.timeline-date {
		color: var(--comment);
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.timeline-card {
		background: var(--current-line);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.3s;
		cursor: pointer;
	}

	.timeline-card:hover {
		transform: translateX(10px);
		box-shadow: 0 3px 10px rgba(0,0,0,0.3);
	}

	.timeline-card .note-title {
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--cyan);
		margin-bottom: 0.5rem;
	}

	.timeline-card .note-content {
		color: var(--foreground);
		font-size: 0.9rem;
		line-height: 1.4;
	}
</style>

<div class="view-controls">
	<div class="view-toggle">
		<button onclick="setView('grid')">Grid</button>
		<button onclick="setView('list')">List</button>
		<button class="active" onclick="setView('timeline')">Timeline</button>
	</div>
	<div class="search-bar">
		<input type="text" placeholder="Search notes..." id="searchInput">
	</div>
	<a href="/users/notes/new" class="btn btn-primary">+ New Note</a>
</div>

<div class="timeline">
	{{range .Notes}}
	<div class="timeline-item">
		<div class="timeline-date">{{.UpdatedAt}}</div>
		<div class="timeline-card" onclick="location.href='/users/notes/{{.ID}}'">
			<div class="note-title">{{.Title}}</div>
			<div class="note-content">{{.Content}}</div>
		</div>
	</div>
	{{end}}
</div>

<button class="fab" onclick="location.href='/users/notes/new'">+</button>

<script>
	function setView(view) {
		location.href = '/users/notes?view=' + view;
	}
</script>`

// Editor view per CLAUDE.md (split markdown/preview)
const editorTemplate = `<style>
	.editor-container {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		height: calc(100vh - 200px);
		margin-top: 1rem;
	}

	.editor-panel {
		background: var(--current-line);
		border-radius: 8px;
		padding: 1rem;
		overflow: auto;
	}

	.editor-panel h3 {
		color: var(--cyan);
		margin-bottom: 1rem;
	}

	#editor {
		width: 100%;
		height: calc(100% - 3rem);
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
		padding: 1rem;
		font-family: 'Courier New', monospace;
		font-size: 1rem;
		resize: none;
	}

	#preview {
		height: calc(100% - 3rem);
		overflow: auto;
		color: var(--foreground);
		line-height: 1.6;
	}

	#preview h1, #preview h2, #preview h3 {
		color: var(--cyan);
		margin-top: 1rem;
		margin-bottom: 0.5rem;
	}

	#preview code {
		background: var(--bg);
		padding: 0.2rem 0.4rem;
		border-radius: 3px;
		font-family: 'Courier New', monospace;
	}

	#preview pre {
		background: var(--bg);
		padding: 1rem;
		border-radius: 5px;
		overflow-x: auto;
	}

	#preview pre code {
		background: none;
		padding: 0;
	}

	.editor-toolbar {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1rem;
		padding: 0.5rem;
		background: var(--bg);
		border-radius: 5px;
	}

	.toolbar-btn {
		padding: 0.5rem 1rem;
		background: var(--current-line);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 3px;
		cursor: pointer;
		font-size: 0.85rem;
	}

	.toolbar-btn:hover {
		background: var(--purple);
		border-color: var(--purple);
	}

	.note-metadata {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.note-metadata input, .note-metadata select {
		width: 100%;
		padding: 0.5rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
	}

	.save-actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
	}
</style>

<div class="note-metadata">
	<input type="text" id="noteTitle" placeholder="Note Title" value="{{.Note.Title}}">
	<select id="noteType">
		<option value="note">📝 Note</option>
		<option value="code">💻 Code</option>
		<option value="checklist">✅ Checklist</option>
		<option value="canvas">🎨 Canvas</option>
		<option value="encrypted">🔒 Encrypted</option>
	</select>
</div>

<div class="editor-toolbar">
	<button class="toolbar-btn" onclick="insertMarkdown('**', '**')"><strong>B</strong></button>
	<button class="toolbar-btn" onclick="insertMarkdown('*', '*')"><em>I</em></button>
	<button class="toolbar-btn" onclick="insertMarkdown('# ', '')">H1</button>
	<button class="toolbar-btn" onclick="insertMarkdown('## ', '')">H2</button>
	<button class="toolbar-btn" onclick="insertMarkdown('- ', '')">List</button>
	<button class="toolbar-btn" onclick="insertMarkdown('` + "`" + "``" + `\n', '\n` + "`" + "``" + `')">Code</button>
	<button class="toolbar-btn" onclick="insertMarkdown('[', '](url)')">Link</button>
</div>

<div class="editor-container">
	<div class="editor-panel">
		<h3>Editor</h3>
		<textarea id="editor" oninput="updatePreview()">{{.Note.Content}}</textarea>
	</div>
	<div class="editor-panel">
		<h3>Preview</h3>
		<div id="preview"></div>
	</div>
</div>

<div class="save-actions">
	<button class="btn btn-primary" onclick="saveNote()">Save</button>
	<button class="btn btn-secondary" onclick="saveDraft()">Save Draft</button>
	<button class="btn btn-danger" onclick="cancelEdit()">Cancel</button>
</div>

<script>
	function updatePreview() {
		const content = document.getElementById('editor').value;
		// Simple markdown parsing (use marked.js in production)
		let html = content
			.replace(/^### (.*$)/gim, '<h3>$1</h3>')
			.replace(/^## (.*$)/gim, '<h2>$1</h2>')
			.replace(/^# (.*$)/gim, '<h1>$1</h1>')
			.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
			.replace(/\*(.*?)\*/g, '<em>$1</em>')
			.replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2">$1</a>')
			.replace(/\n/g, '<br>');
		document.getElementById('preview').innerHTML = html;
	}

	function insertMarkdown(before, after) {
		const editor = document.getElementById('editor');
		const start = editor.selectionStart;
		const end = editor.selectionEnd;
		const text = editor.value;
		const selected = text.substring(start, end);
		editor.value = text.substring(0, start) + before + selected + after + text.substring(end);
		editor.focus();
		updatePreview();
	}

	function saveNote() {
		const data = {
			title: document.getElementById('noteTitle').value,
			content: document.getElementById('editor').value,
			note_type: document.getElementById('noteType').value
		};
		fetch('/api/v1/notes/{{.Note.ID}}', {
			method: 'PUT',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify(data)
		}).then(() => location.href = '/users/notes');
	}

	function saveDraft() {
		localStorage.setItem('draft_{{.Note.ID}}', document.getElementById('editor').value);
		alert('Draft saved locally');
	}

	function cancelEdit() {
		if (confirm('Discard changes?')) {
			location.href = '/users/notes';
		}
	}

	// Load draft if exists
	const draft = localStorage.getItem('draft_{{.Note.ID}}');
	if (draft && !document.getElementById('editor').value) {
		document.getElementById('editor').value = draft;
	}

	updatePreview();

	// Auto-save per CLAUDE.md (30 seconds)
	setInterval(saveDraft, 30000);
</script>`

// Login page
const loginTemplate = `<style>
	.auth-container {
		max-width: 400px;
		margin: 4rem auto;
		background: var(--current-line);
		padding: 2rem;
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.3);
	}

	.auth-container h2 {
		color: var(--green);
		text-align: center;
		margin-bottom: 2rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		color: var(--foreground);
	}

	.form-group input {
		width: 100%;
		padding: 0.75rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
		font-size: 1rem;
	}

	.form-group input:focus {
		outline: none;
		border-color: var(--cyan);
	}

	.form-actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 2rem;
	}

	.form-actions button {
		width: 100%;
		padding: 0.75rem;
		font-size: 1rem;
	}

	.form-footer {
		margin-top: 1.5rem;
		text-align: center;
		color: var(--comment);
	}

	.form-footer a {
		color: var(--cyan);
		text-decoration: none;
	}

	.form-footer a:hover {
		color: var(--green);
	}
</style>

<div class="auth-container">
	<h2>Login to casnotes</h2>

	{{if .Error}}
	<div class="alert alert-error">{{.Error}}</div>
	{{end}}

	<form method="POST" action="/login">
		<div class="form-group">
			<label for="username">Username or Email</label>
			<input type="text" id="username" name="username" required>
		</div>
		<div class="form-group">
			<label for="password">Password</label>
			<input type="password" id="password" name="password" required>
		</div>
		<div class="form-group">
			<label>
				<input type="checkbox" name="remember" value="true">
				Remember me for 90 days
			</label>
		</div>
		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Login</button>
		</div>
	</form>

	<div class="form-footer">
		<p>Don't have an account? <a href="/register">Register</a></p>
		<p><a href="/forgot-password">Forgot password?</a></p>
	</div>
</div>`

// Register page
const registerTemplate = `<style>
	.auth-container {
		max-width: 400px;
		margin: 4rem auto;
		background: var(--current-line);
		padding: 2rem;
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.3);
	}

	.auth-container h2 {
		color: var(--green);
		text-align: center;
		margin-bottom: 2rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		color: var(--foreground);
	}

	.form-group input {
		width: 100%;
		padding: 0.75rem;
		background: var(--bg);
		color: var(--foreground);
		border: 1px solid var(--comment);
		border-radius: 5px;
		font-size: 1rem;
	}

	.form-group input:focus {
		outline: none;
		border-color: var(--cyan);
	}

	.form-actions {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 2rem;
	}

	.form-actions button {
		width: 100%;
		padding: 0.75rem;
		font-size: 1rem;
	}

	.form-footer {
		margin-top: 1.5rem;
		text-align: center;
		color: var(--comment);
	}

	.form-footer a {
		color: var(--cyan);
		text-decoration: none;
	}

	.form-footer a:hover {
		color: var(--green);
	}
</style>

<div class="auth-container">
	<h2>Register for casnotes</h2>

	{{if .Error}}
	<div class="alert alert-error">{{.Error}}</div>
	{{end}}

	<form method="POST" action="/register">
		<div class="form-group">
			<label for="username">Username</label>
			<input type="text" id="username" name="username" required>
		</div>
		<div class="form-group">
			<label for="email">Email</label>
			<input type="email" id="email" name="email" required>
		</div>
		<div class="form-group">
			<label for="password">Password (min 8 characters)</label>
			<input type="password" id="password" name="password" required minlength="8">
		</div>
		<div class="form-group">
			<label for="confirm_password">Confirm Password</label>
			<input type="password" id="confirm_password" name="confirm_password" required>
		</div>
		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Register</button>
		</div>
	</form>

	<div class="form-footer">
		<p>Already have an account? <a href="/login">Login</a></p>
	</div>
</div>`

// Index/landing page
const indexTemplate = `<style>
	.hero {
		text-align: center;
		padding: 4rem 2rem;
	}

	.hero h1 {
		font-size: 3rem;
		color: var(--green);
		margin-bottom: 1rem;
	}

	.hero p {
		font-size: 1.5rem;
		color: var(--comment);
		margin-bottom: 2rem;
	}

	.hero-actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}

	.features {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 2rem;
		margin: 4rem 0;
	}

	.feature-card {
		background: var(--current-line);
		padding: 2rem;
		border-radius: 8px;
		text-align: center;
	}

	.feature-card .icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.feature-card h3 {
		color: var(--cyan);
		margin-bottom: 1rem;
	}

	.feature-card p {
		color: var(--comment);
	}
</style>

<div class="hero">
	<h1>📝 Welcome to casnotes</h1>
	<p>Self-hosted, Git-powered note-taking</p>
	<div class="hero-actions">
		<a href="/register" class="btn btn-primary">Get Started</a>
		<a href="/discover" class="btn btn-secondary">Discover Public Notes</a>
	</div>
</div>

<div class="features">
	<div class="feature-card">
		<div class="icon">🔒</div>
		<h3>Private & Secure</h3>
		<p>Your data, your server. Full control over your notes with encryption support.</p>
	</div>
	<div class="feature-card">
		<div class="icon">🎨</div>
		<h3>Multiple Note Types</h3>
		<p>Markdown, code snippets, checklists, canvas drawings, and encrypted notes.</p>
	</div>
	<div class="feature-card">
		<div class="icon">📁</div>
		<h3>Organized</h3>
		<p>Notebooks, tags, pinning, and smart collections to keep everything tidy.</p>
	</div>
	<div class="feature-card">
		<div class="icon">🔍</div>
		<h3>Powerful Search</h3>
		<p>Full-text search with SQLite FTS5 to find your notes instantly.</p>
	</div>
	<div class="feature-card">
		<div class="icon">🌙</div>
		<h3>Beautiful Themes</h3>
		<p>Dark (Dracula) and Light (GitHub) themes with auto-detection.</p>
	</div>
	<div class="feature-card">
		<div class="icon">⚡</div>
		<h3>Fast & Lightweight</h3>
		<p>Single static binary, zero dependencies, instant deployment.</p>
	</div>
</div>`
