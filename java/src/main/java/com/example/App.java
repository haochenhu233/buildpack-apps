package com.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@SpringBootApplication
@RestController
public class App {
    public static void main(String[] args) {
        // Spring Boot binds server.port=${PORT} automatically on CF.
        // Do NOT hardcode server.port — that broke on jbp 5.0.x (no more auto-reconfiguration).
        SpringApplication.run(App.class, args);
    }

    @GetMapping("/")
    public Map<String, Object> root() {
        return Map.of(
            "runtime", "java " + System.getProperty("java.version"),
            "vendor", System.getProperty("java.vendor"),
            "ok", true);
    }
}
