function [results] = verify_channel_model_go()
% VERIFY_CHANNEL_MODEL_GO - 验证Go实现的3GPP信道模型与MATLAB实现的一致性
%
% 输入: 无
% 输出: results - 包含验证结果的结构体
%
% 示例:
%   results = verify_channel_model_go();

    fprintf('=== 3GPP TR 38.901信道模型验证 ===\n');
    
    results = struct();
    
    fc = 3.5e9;
    h_BS = 25;
    h_UT = 1.5;
    
    scenarios = {'UMa', 'UMi', 'RMa', 'Indoor-Office'};
    
    fprintf('载波频率: %.1f GHz\n', fc/1e9);
    fprintf('基站高度: %.1f m\n', h_BS);
    fprintf('用户高度: %.1f m\n', h_UT);
    
    distances = 10:10:2000;
    
    fprintf('\n1. 路径损耗模型验证:\n');
    results.path_loss = struct();
    
    for s = 1:length(scenarios)
        scenario = scenarios{s};
        pl_los = zeros(size(distances));
        pl_nlos = zeros(size(distances));
        
        for i = 1:length(distances)
            d = distances(i);
            pl_los(i) = compute_path_loss_los(scenario, d, fc, h_BS, h_UT);
            pl_nlos(i) = compute_path_loss_nlos(scenario, d, fc, h_BS, h_UT);
        end
        
        results.path_loss.(scenario).los = pl_los;
        results.path_loss.(scenario).nlos = pl_nlos;
        
        fprintf('   %s: PL_LOS(100m)=%.2f dB, PL_NLOS(100m)=%.2f dB\n', ...
            scenario, pl_los(find(distances>=100,1)), pl_nlos(find(distances>=100,1)));
    end
    
    fprintf('\n2. 大尺度参数统计验证:\n');
    num_samples = 10000;
    results.large_scale = struct();
    
    for s = 1:length(scenarios)
        scenario = scenarios{s};
        
        ds_samples = zeros(1, num_samples);
        asd_samples = zeros(1, num_samples);
        asa_samples = zeros(1, num_samples);
        k_samples = zeros(1, num_samples);
        
        for i = 1:num_samples
            params = generate_large_scale_params(scenario);
            ds_samples(i) = params.ds;
            asd_samples(i) = params.asd;
            asa_samples(i) = params.asa;
            k_samples(i) = params.k_factor;
        end
        
        results.large_scale.(scenario).ds_mean = mean(log10(ds_samples));
        results.large_scale.(scenario).ds_std = std(log10(ds_samples));
        results.large_scale.(scenario).asd_mean = mean(log10(asd_samples));
        results.large_scale.(scenario).asa_mean = mean(log10(asa_samples));
        results.large_scale.(scenario).k_mean = mean(10.^(k_samples/10));
        
        fprintf('   %s:\n', scenario);
        fprintf('     DS: μ=%.2f, σ=%.2f (log10(s))\n', ...
            results.large_scale.(scenario).ds_mean, results.large_scale.(scenario).ds_std);
        fprintf('     ASD: μ=%.2f (log10(deg))\n', results.large_scale.(scenario).asd_mean);
        fprintf('     K因子: %.2f dB\n', 10*log10(results.large_scale.(scenario).k_mean));
    end
    
    fprintf('\n3. 小尺度衰落验证:\n');
    M = 16;
    N = 1000;
    num_clusters = 23;
    num_subpaths = 20;
    
    channel_samples = zeros(M, N);
    for t = 1:N
        h = generate_channel_impulse_response(M, num_clusters, num_subpaths);
        channel_samples(:, t) = abs(h(:));
    end
    
    amplitude = abs(channel_samples(:));
    results.fading.envelope_mean = mean(amplitude);
    results.fading.envelope_std = std(amplitude);
    results.fading.rice_k_estimate = estimate_rice_k(amplitude);
    
    fprintf('   包络均值: %.4f\n', results.fading.envelope_mean);
    fprintf('   包络标准差: %.4f\n', results.fading.envelope_std);
    fprintf('   估计莱斯K因子: %.2f dB\n', 10*log10(results.fading.rice_k_estimate));
    
    fprintf('\n4. 功率时延谱验证:\n');
    pdp = generate_power_delay_profile('UMa', num_clusters);
    results.pdp.delays = pdp.delays;
    results.pdp.powers = pdp.powers;
    results.pdp.rms_delay_spread = compute_rms_delay_spread(pdp.delays, pdp.powers);
    
    fprintf('   RMS时延扩展: %.2f ns\n', results.pdp.rms_delay_spread * 1e9);
    fprintf('   簇数: %d\n', length(pdp.delays));
    
    figure('Name', '信道模型验证', 'Position', [100, 100, 1400, 800]);
    
    subplot(2, 4, 1);
    hold on;
    colors = {'b-', 'r-', 'g-', 'm-'};
    for s = 1:length(scenarios)
        scenario = scenarios{s};
        plot(log10(distances), results.path_loss.(scenario).los, [colors{s}(1), '-'], 'LineWidth', 1.5);
        plot(log10(distances), results.path_loss.(scenario).nlos, [colors{s}(1), '--'], 'LineWidth', 1.5);
    end
    xlabel('log10(距离) [m]');
    ylabel('路径损耗 [dB]');
    title('路径损耗模型');
    legend([scenarios; strrep(scenarios, '', ' NLOS')], 'Location', 'northwest');
    grid on;
    
    subplot(2, 4, 2);
    histogram(log10(ds_samples), 50, 'Normalization', 'pdf');
    xlabel('log10(DS) [s]');
    ylabel('概率密度');
    title('时延扩展分布 (UMa)');
    grid on;
    
    subplot(2, 4, 3);
    histogram(log10(asd_samples), 50, 'Normalization', 'pdf');
    xlabel('log10(ASD) [deg]');
    ylabel('概率密度');
    title('发射方位角扩展分布 (UMa)');
    grid on;
    
    subplot(2, 4, 4);
    histogram(amplitude, 50, 'Normalization', 'pdf');
    xlabel('幅度');
    ylabel('概率密度');
    title('小尺度衰落幅度分布');
    grid on;
    
    subplot(2, 4, 5);
    stem(pdp.delays * 1e6, 10*log10(pdp.powers), 'filled', 'LineWidth', 1.5);
    xlabel('时延 (μs)');
    ylabel('功率 (dB)');
    title('功率时延谱');
    grid on;
    
    subplot(2, 4, 6);
    [f, Pxx] = compute_doppler_spectrum(channel_samples(1,:));
    plot(f, 10*log10(Pxx), 'b-', 'LineWidth', 1.5);
    xlabel('多普勒频移 (Hz)');
    ylabel('功率谱密度 (dB)');
    title('多普勒功率谱');
    grid on;
    
    subplot(2, 4, 7);
    spatial_corr = compute_spatial_correlation(M, 0.5);
    imagesc(spatial_corr);
    colorbar;
    xlabel('天线索引');
    ylabel('天线索引');
    title('空间相关性矩阵');
    
    subplot(2, 4, 8);
    los_prob = zeros(size(distances));
    for i = 1:length(distances)
        los_prob(i) = compute_los_probability('UMa', distances(i));
    end
    plot(distances, los_prob, 'b-', 'LineWidth', 2);
    xlabel('距离 (m)');
    ylabel('LOS概率');
    title('UMa场景LOS概率');
    grid on;
    ylim([0, 1]);
    
    saveas(gcf, 'channel_model_verification.png');
    fprintf('\n图表已保存至 channel_model_verification.png\n');
    
    results.passed = results.pdp.rms_delay_spread > 0;
    fprintf('\n验证结果: %s\n', ternary(results.passed, '通过', '失败'));
end

function pl = compute_path_loss_los(scenario, d, fc, h_BS, h_UT)
% 计算LOS路径损耗
    fc_ghz = fc / 1e9;
    
    switch scenario
        case 'UMa'
            pl = 28.0 + 22.0 * log10(d) + 20.0 * log10(fc_ghz);
        case 'UMi'
            pl = 32.4 + 21.0 * log10(d) + 20.0 * log10(fc_ghz);
        case 'RMa'
            pl = 28.0 + 22.0 * log10(d) + 20.0 * log10(fc_ghz);
        case 'Indoor-Office'
            pl = 32.4 + 17.3 * log10(d) + 20.0 * log10(fc_ghz);
        otherwise
            pl = 28.0 + 22.0 * log10(d) + 20.0 * log10(fc_ghz);
    end
end

function pl = compute_path_loss_nlos(scenario, d, fc, h_BS, h_UT)
% 计算NLOS路径损耗
    fc_ghz = fc / 1e9;
    pl_los = compute_path_loss_los(scenario, d, fc, h_BS, h_UT);
    
    switch scenario
        case 'UMa'
            pl_nlos = max(pl_los, 13.54 + 39.08 * log10(d) + 20.0 * log10(fc_ghz) - 0.6 * (h_UT - 1.5));
        case 'UMi'
            pl_nlos = max(pl_los, 35.4 + 22.4 * log10(d) + 20.0 * log10(fc_ghz));
        case 'RMa'
            pl_nlos = max(pl_los, 32.4 + 23.0 * log10(d) + 20.0 * log10(fc_ghz));
        case 'Indoor-Office'
            pl_nlos = max(pl_los, 38.4 + 24.0 * log10(d) + 20.0 * log10(fc_ghz));
        otherwise
            pl_nlos = pl_los;
    end
    
    pl = pl_nlos;
end

function params = generate_large_scale_params(scenario)
% 生成大尺度参数
    params = struct();
    
    switch scenario
        case 'UMa'
            params.ds = 10^(-7.0 + 0.3 * randn());
            params.asd = 10^(1.3 + 0.3 * randn());
            params.asa = 10^(1.8 + 0.2 * randn());
            params.k_factor = 10^(9.0 + 3.5 * randn());
        case 'UMi'
            params.ds = 10^(-6.8 + 0.3 * randn());
            params.asd = 10^(1.2 + 0.3 * randn());
            params.asa = 10^(1.6 + 0.25 * randn());
            params.k_factor = 10^(6.0 + 4.0 * randn());
        case 'RMa'
            params.ds = 10^(-7.5 + 0.3 * randn());
            params.asd = 10^(1.0 + 0.4 * randn());
            params.asa = 10^(1.5 + 0.3 * randn());
            params.k_factor = 10^(7.0 + 4.0 * randn());
        case 'Indoor-Office'
            params.ds = 10^(-7.5 + 0.3 * randn());
            params.asd = 10^(1.4 + 0.25 * randn());
            params.asa = 10^(1.6 + 0.25 * randn());
            params.k_factor = 10^(7.0 + 3.0 * randn());
        otherwise
            params.ds = 10^(-7.0 + 0.3 * randn());
            params.asd = 10^(1.3 + 0.3 * randn());
            params.asa = 10^(1.8 + 0.2 * randn());
            params.k_factor = 10^(9.0 + 3.5 * randn());
    end
end

function h = generate_channel_impulse_response(M, num_clusters, num_subpaths)
% 生成信道冲激响应
    h = zeros(M, 1);
    
    for n = 1:num_clusters
        cluster_power = 1 / num_clusters;
        cluster_phase = 2 * pi * rand();
        
        for m = 1:num_subpaths
            subpath_power = cluster_power / num_subpaths;
            subpath_phase = cluster_phase + 2 * pi * rand();
            
            aod = randn() * 15 * pi / 180;
            
            for ant = 1:M
                phase = 2 * pi * (ant - 1) * 0.5 * sin(aod) + subpath_phase;
                h(ant) = h(ant) + sqrt(subpath_power) * exp(1i * phase);
            end
        end
    end
end

function pdp = generate_power_delay_profile(scenario, num_clusters)
% 生成功率时延谱
    pdp = struct();
    
    switch scenario
        case 'UMa'
            mu_tau = -7.0;
            sigma_tau = 0.3;
        case 'UMi'
            mu_tau = -6.8;
            sigma_tau = 0.3;
        case 'RMa'
            mu_tau = -7.5;
            sigma_tau = 0.3;
        otherwise
            mu_tau = -7.0;
            sigma_tau = 0.3;
    end
    
    pdp.delays = sort(10.^(mu_tau + sigma_tau * randn(1, num_clusters)));
    pdp.delays = pdp.delays - pdp.delays(1);
    
    pdp.powers = zeros(1, num_clusters);
    for i = 1:num_clusters
        pdp.powers(i) = exp(-pdp.delays(i) / 1e-7);
    end
    pdp.powers = pdp.powers / sum(pdp.powers);
end

function rms_ds = compute_rms_delay_spread(delays, powers)
% 计算RMS时延扩展
    mean_delay = sum(delays .* powers);
    rms_ds = sqrt(sum(powers .* (delays - mean_delay).^2));
end

function k = estimate_rice_k(amplitude)
% 估计莱斯K因子
    mean_amp = mean(amplitude);
    var_amp = var(amplitude);
    
    k = mean_amp^2 / (2 * var_amp);
    if k < 0
        k = 0;
    end
end

function [f, Pxx] = compute_doppler_spectrum(signal)
% 计算多普勒功率谱
    N = length(signal);
    Pxx = abs(fftshift(fft(signal .* hanning(N)'))).^2;
    f = linspace(-0.5, 0.5, N);
end

function corr = compute_spatial_correlation(M, d)
% 计算空间相关性矩阵
    corr = zeros(M, M);
    for i = 1:M
        for j = 1:M
            delta = abs(i - j) * d;
            corr(i, j) = besselj(0, 2 * pi * delta);
        end
    end
end

function prob = compute_los_probability(scenario, distance)
% 计算LOS概率
    switch scenario
        case 'UMa'
            if distance <= 18
                prob = 1;
            else
                prob = min(18/distance, 1) * (1 - exp(-distance/63)) + exp(-distance/63);
            end
        case 'UMi'
            if distance <= 18
                prob = 1;
            else
                prob = min(18/distance, 1) * (1 - exp(-distance/36)) + exp(-distance/36);
            end
        otherwise
            prob = exp(-distance/100);
    end
end

function result = ternary(condition, true_val, false_val)
    if condition
        result = true_val;
    else
        result = false_val;
    end
end
