<script lang="ts">
    import { onMount, tick } from "svelte";
    import "../app.css";

    let messages = $state<string[]>([]);
    let ws = $state<WebSocket | null>(null);
    let readyState = $state<number | null>(null);

    let message = $state("");
    
    // Auto-reconnect configuration
    let reconnectAttempts = $state(0);
    let reconnectInterval = $state(1000); // Start with 1s interval
    let maxReconnectInterval = 30000; // Max 30s between attempts
    let reconnectTimer: number | null = $state(null);
    let isReconnecting = $state(false);

    let messageContainer: HTMLDivElement | null = null;
    let userScrolledUp = $state(false);
    const scrollThreshold = 50;

    // --- Dark Mode State ---
    let isDarkMode = $state(false); // Default to light mode

    function toggleDarkMode() {
        isDarkMode = !isDarkMode;
        
        if (typeof document !== 'undefined') {
            if (isDarkMode) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
            localStorage.setItem('theme', isDarkMode ? 'dark' : 'light');
        }
    }

    // Optional: Map numbers to strings for better display
    const stateMap: { [key: number]: string } = {
        0: "CONNECTING",
        1: "OPEN",
        2: "CLOSING",
        3: "CLOSED"
    };
    let readyStateString: string = $derived(readyState !== null ? (stateMap[readyState] || "UNKNOWN") : "DISCONNECTED");

    function isNearBottom(element: HTMLDivElement | null): boolean {
        if (!element) return true;
        return element.scrollHeight - element.scrollTop - element.clientHeight < scrollThreshold;
    }

    function handleScroll() {
        if (!messageContainer) return;
        if (isNearBottom(messageContainer)) {
            userScrolledUp = false;
        } else {
            userScrolledUp = true;
        }
    }

    const sendMessage = async () => {
        const formData = new FormData();
        formData.append("message", message);
        try {
            const result = await fetch("http://localhost:32768/chat", {
                method: "POST",
                body: formData
            });
            if (result.ok) {
                message = ""; // Clear input only on success
            } else {
                console.error("Failed to send message:", result.status);
                // Handle error display if needed
            }
        } catch (error) {
            console.error("Error sending message:", error);
        }
    }

    function connectWebSocket() {
        if (isReconnecting) return;
        
        const socket = new WebSocket("ws://localhost:32768/ws");
        ws = socket;
        readyState = socket.readyState;

        socket.onopen = () => {
            console.log("Connected to server");
            readyState = socket.readyState;
            reconnectAttempts = 0;
            reconnectInterval = 1000;
            isReconnecting = false;
        };

        socket.onmessage = async (event) => {
            console.log("Message from server ", event.data);

            const shouldScroll = isNearBottom(messageContainer);

            messages.push(event.data);
            messages = messages;

            await tick();

            if (shouldScroll && !userScrolledUp && messageContainer) {
                messageContainer.scrollTop = messageContainer.scrollHeight;
            }
        };

        socket.onclose = (event) => {
            console.log("Disconnected from server:", event.code, event.reason);
            readyState = socket.readyState;
            ws = null;
            
            // Only attempt to reconnect on abnormal closure or if not deliberately closed
            if (event.code !== 1000 && event.code !== 1001) {
                scheduleReconnect();
            }
        };

        socket.onerror = (error) => {
            console.error("WebSocket Error:", error);
            readyState = socket.readyState;
        };
    }

    function scheduleReconnect() {
        if (reconnectTimer !== null) {
            window.clearTimeout(reconnectTimer);
        }
        
        isReconnecting = true;
        
        // Use exponential backoff with jitter for reconnection
        const jitter = Math.random() * 0.5 + 0.75; // Random value between 0.75 and 1.25
        const delay = Math.min(reconnectInterval * jitter, maxReconnectInterval);
        
        console.log(`Attempting to reconnect in ${Math.round(delay)}ms (attempt ${reconnectAttempts + 1})`);
        
        reconnectTimer = window.setTimeout(() => {
            reconnectAttempts++;
            reconnectInterval = Math.min(reconnectInterval * 2, maxReconnectInterval);
            connectWebSocket();
        }, delay);
    }

    onMount(() => {
        // Check for dark mode preference
        const savedTheme = localStorage.getItem('theme');
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        
        // Set initial state based on saved preference or system preference
        isDarkMode = savedTheme ? savedTheme === 'dark' : prefersDark;
        
        // Apply dark mode class based on initial state
        if (isDarkMode) {
            document.documentElement.classList.add('dark');
        }

        console.log("Mounted");
        connectWebSocket();

        const fetchMessages = async () => {
            try {
                const result = await fetch("http://localhost:32768/chat");
                if (result.ok) {
                    const data = await result.json();
                    if (Array.isArray(data)) {
                        messages = data.map((msg: any) => msg.Content).reverse();
                    } else {
                        console.error("Fetched data is not an array:", data);
                        messages = []; // Reset or handle as appropriate
                    }
                    // Scroll after initial fetch
                    await tick();
                    if (messageContainer) {
                        messageContainer.scrollTop = messageContainer.scrollHeight;
                    }
                    userScrolledUp = false;
                } else {
                    console.error("Failed to fetch initial messages:", result.status);
                }
            } catch(error) {
                 console.error("Error fetching initial messages:", error);
            }
        }

        fetchMessages();

        // Cleanup on component unmount
        return () => {
            if (ws && ws.readyState < WebSocket.CLOSING) {
                console.log("Closing WebSocket connection on unmount");
                ws.close(1000, "Component unmounted");
            }
            
            if (reconnectTimer !== null) {
                window.clearTimeout(reconnectTimer);
                reconnectTimer = null;
            }
            
            ws = null;
            readyState = null;
        };
    });
</script>

<!-- Update the root element without conditional dark class -->
<div class="flex flex-col h-screen p-4">
    <div class="bg-gray-50 dark:bg-gray-900 flex flex-col h-full rounded shadow">

        <!-- Header Area -->
        <div class="mb-4 p-2 bg-white dark:bg-gray-800 rounded-t shadow flex justify-between items-center">
            <div>
                <p class="text-sm font-medium text-gray-500 dark:text-gray-400">CONNECTION:</p>
                <p class="font-semibold" class:text-green-600={readyState === 1}
                   class:text-red-600={readyState !== 1}
                   class:dark:text-green-400={readyState === 1}
                   class:dark:text-red-400={readyState !== 1}>
                    {readyStateString}
                </p>
            </div>
            <!-- Dark Mode Toggle Button -->
            <button
                onclick={toggleDarkMode}
                class="p-2 rounded bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 transition-colors duration-200"
                aria-label="Toggle dark mode"
            >
                {isDarkMode ? '☀️' : '🌙'} <!-- Icons for toggle -->
            </button>
        </div>

        <h1 class="text-2xl font-bold mb-4 px-4 text-gray-800 dark:text-gray-100">Simple Chat</h1>

        <!-- Message Display Area -->
        <div
            bind:this={messageContainer}
            onscroll={handleScroll}
            class="flex-grow bg-white dark:bg-gray-800 overflow-y-auto p-4 mb-4 mx-4 border border-gray-200 dark:border-gray-600 rounded shadow-sm dark:shadow-md scroll-smooth"
        >
            <ul class="space-y-2">
                {#each messages as msg, i (i)}
                    <li class="p-2 rounded break-words bg-blue-50 dark:bg-gray-700 text-gray-700 dark:text-gray-200">
                        {msg}
                    </li>
                {/each}
                {#if messages.length === 0}
                    <li class="text-gray-400 dark:text-gray-500 italic">No messages yet...</li>
                {/if}
            </ul>
        </div>

        <!-- Input Area -->
        <div class="mt-auto flex items-center space-x-2 p-2 bg-gray-100 dark:bg-gray-800 rounded-b shadow">
            <input
                type="text"
                bind:value={message}
                placeholder="Type your message..."
                class="flex-grow p-2 border border-gray-300 dark:border-gray-600 rounded shadow-sm focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 transition-colors duration-200"
                onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter') sendMessage(); }}
            />
            <button
                onclick={sendMessage}
                disabled={readyState !== 1 || !message.trim()}
                class="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white font-semibold rounded shadow hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 disabled:opacity-50 disabled:bg-gray-400 dark:disabled:bg-gray-600 disabled:cursor-not-allowed transition-colors duration-200"
            >
                Send
            </button>
        </div>
    </div>
</div>