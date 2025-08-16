<script lang="ts">
    import { EventsOn } from "../wailsjs/runtime/runtime";

    let predictionText = "";
    let bufferText = "";
    let correctedText = "";

    EventsOn("setPredictionText", (text: string) => {
        predictionText = text;
    });

    EventsOn("setBufferText", (text: string) => {
        bufferText = text;
    });

    EventsOn("setCorrectedText", (text: string) => {
        correctedText = text;
    });
</script>

<main>
    <div class="bufferText" id="result">
        <b>Buffer:</b>
        <br />
        {bufferText}
    </div>

    <br />

    <div class="predictionText" id="result">
        <b>Prediction:</b>
        <br />
        {predictionText}
    </div>

    <br />

    <div class="predictionText" id="result">
        <b>Corrected:</b>
        <br />
        {correctedText}
    </div>
</main>
