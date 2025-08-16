<script lang="ts">
    import { EventsOn } from "../wailsjs/runtime/runtime";

    let predictionText = "";
    let bufferText = "";

    EventsOn("setPredictionText", (text: string) => {
        predictionText = text;
    });

    EventsOn("setBufferText", (text: string) => {
        bufferText = text;
    });
</script>

<main>
    <div class="bufferText" id="result">{bufferText}</div>

    <div class="predictionText" id="result">{predictionText}</div>
</main>
