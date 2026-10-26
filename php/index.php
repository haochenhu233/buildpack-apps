<?php
header('Content-Type: application/json');
echo json_encode(["runtime" => "php " . phpversion(), "ok" => true]);
